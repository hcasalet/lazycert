package lc

import (
	"errors"
	hm "github.com/cornelk/hashmap"
	"golang.org/x/net/context"
	"log"
	"time"
)

type EdgeService struct {
	key                *Key
	log                *Log
	leader             *LeaderConfig
	currentTerm        int32
	config             *Config
	teClient           *TEClient
	clusterClient      *EdgeClient
	regConfig          *RegistrationConfig
	lc                 *LeaderClient
	iAmLeader          bool
	queue              *TimedQueue
	newLogEntryChannel chan *LogEntry
	certificationTimes hm.HashMap
}

func NewEdgeService(configuration *Config) *EdgeService {
	key := NewKey(configuration.PrivateKeyFileName)
	tec := NewTEClient(configuration.TEAddr)
	edgeService := &EdgeService{
		key:                key,
		leader:             nil,
		currentTerm:        0,
		config:             configuration,
		teClient:           tec,
		lc:                 nil,
		iAmLeader:          false,
		log:                NewLog(configuration),
		newLogEntryChannel: make(chan *LogEntry),
		certificationTimes: hm.HashMap{},
	}
	//edgeService.log.SetLogEntryUpdateChannel(edgeService.newLogEntryChannel)
	edgeService.queue = NewTimedQueue(configuration.Epoch.Duration, configuration.Epoch.MaxSize, edgeService.log.BatchedData)
	go edgeService.waitForLogEntryUpdate()
	log.Printf("\nSetting up connection to edge replicas.\n")
	time.Sleep(time.Second * 10)
	edgeService.clusterClient = createNewEdgeClient(configuration)
	log.Printf("\nConnection to edge replicas setup.\n")
	return edgeService
}

func (e *EdgeService) Commit(ctx context.Context, commitData *CommitData) (*Dummy, error) {
	//log.Printf("Received data to commit: %v", commitData)
	if e.iAmLeader {
		//log.Printf("Inserting data into commit queue.")
		e.queue.Insert(commitData)
	} else {
		e.lc.sendCommitDataToLeader(commitData)
	}
	return &Dummy{}, nil
}

func (e *EdgeService) Read(ctx context.Context, val *KeyVal) (*ReadResponse, error) {
	startTime := time.Now()
	var rr *ReadResponse
	key := string(val.Key)
	if v, ok := e.log.Read(key); ok {
		rr = &ReadResponse{
			Data: &KeyVal{
				Key:   val.Key,
				Value: v,
			}, Status: ResponseStatus_SUCCESS}
	} else {
		rr = &ReadResponse{Data: &KeyVal{Key: val.Key, Value: []byte{0}}, Status: ResponseStatus_ERROR}
	}
	log.Printf("READ COMPLETED IN: %s, %s", time.Since(startTime), rr.Status)
	return rr, nil
}

func (e *EdgeService) Propose(ctx context.Context, data *ProposeData) (d *Dummy, err error) {
	// TODO: validate header to ensure data was received from the right leader.
	if e.log.Propose(data.LogBlock.LogID, data.LogBlock.Data) {
		startTime := time.Now()
		m, _ := e.log.logEntry.Get(data.LogBlock.LogID)
		entry := m.(*LogEntry)
		a := ConvertToAcceptMsg(entry, &e.config.Node, e.currentTerm, e.key)
		e.teClient.SendAccept(a)
		e.certificationTimes.Set(data.LogBlock.LogID, startTime)
		err = nil
	} else {
		err = errors.New("cannot add duplicate propose data")
	}
	return &Dummy{}, err
}

func (e *EdgeService) HeartBeat(ctx context.Context, info *HeartBeatInfo) (*Dummy, error) {
	panic("implement me")
}

func (e *EdgeService) Certification(ctx context.Context, certificate *Certificate) (d *Dummy, err error) {
	if VerifyMessage(
		certificate.AcceptHash,
		certificate.TeSignature,
		GetPublicKeyFromBytes(e.regConfig.TePublicKey.RawPublicKey)) {
		//log.Printf("Verfied certificate received from TE")
		e.log.Certificate <- certificate
		err = nil
		if t, ok := e.certificationTimes.Get(certificate.LogID); ok {
			startTime := t.(time.Time)
			log.Printf(" LOG ID, CERTIFICATION DURATION:%v, %s", certificate.LogID, time.Since(startTime))
		} else {
			log.Printf("CERTIFICATION START TIME NOT FOUND FOR LOG ID: %v", certificate.LogID)
		}
	} else {
		log.Printf("Certificate could not be verified. Signature does not match that of TE.")
		err = errors.New("signature match failed. invalid TE signature")
	}
	return &Dummy{}, err
}

func (e *EdgeService) LeaderStatus(ctx context.Context, leaderConfig *LeaderConfig) (*Dummy, error) {
	e.leader = leaderConfig
	e.currentTerm = leaderConfig.TermID
	e.checkLeadershipStatusAndConnect(leaderConfig)
	return &Dummy{}, nil
}

func (e *EdgeService) checkLeadershipStatusAndConnect(leaderConfig *LeaderConfig) {
	log.Printf("Received leader configuration: %v, TermID: %v", leaderConfig.Node, leaderConfig.TermID)
	if e.config.Node.Uuid == leaderConfig.Node.Uuid {
		log.Printf("Now I am the leader.")
		e.iAmLeader = true
		e.log.SetLogEntryUpdateChannel(e.newLogEntryChannel)
	}
	if !e.iAmLeader {
		e.log.SetLogEntryUpdateChannel(nil)
		if e.lc == nil {
			e.lc = NewLeaderClient()
		}
		go e.lc.ConnectToLeader(e.leader.Node)
	}
}

func (e *EdgeService) RegisterWithTE() {
	regConfig, err := e.teClient.Register(PublicKey{
		RawPublicKey: e.key.GetPublicKey(),
	}, e.config.Node, e.currentTerm)
	if err == nil {
		log.Printf("Registered with TE.")
		e.currentTerm = regConfig.ClusterLeader.TermID
		if regConfig.ClusterLeader.Node != nil {
			e.leader = regConfig.ClusterLeader
			e.checkLeadershipStatusAndConnect(regConfig.ClusterLeader)
		} else {
			log.Printf("Leader does not exist. Performing self promotion.")
			go e.teClient.SendSelfPromote(e.config.Node, e.currentTerm+1, e.key.GetPublicKey())
		}
		e.regConfig = regConfig
	} else {
		log.Printf("Error while registering with TE: %v", err)
	}
}

func (e *EdgeService) waitForLogEntryUpdate() {
	for l := range e.newLogEntryChannel {
		startTime := time.Now()
		go e.teClient.SendAccept(ConvertToAcceptMsg(l, &e.config.Node, e.currentTerm, e.key))
		e.certificationTimes.Set(l.LogID, startTime)
		p := ConvertToProposeData(*l, &e.config.Node)
		e.clusterClient.SendProposal(p)
		log.Printf("LOG REPLICATION TIME AT LEADER: %s", time.Since(startTime))
	}
}

func createNewEdgeClient(config *Config) *EdgeClient {
	clusterClient := NewEdgeClient()
	for _, c := range config.ClusterNodes {
		addr := c.Ip + ":" + c.Port
		clusterClient.AddConnection(addr)
	}
	return clusterClient
}
