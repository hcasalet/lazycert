package lc

import (
	"errors"
	"golang.org/x/net/context"
	"log"
)

type EdgeService struct {
	key                *Key
	log                *Log
	leader             *LeaderConfig
	currentTerm        int32
	config             *Config
	teClient           *TEClient
	regConfig          *RegistrationConfig
	lc                 *LeaderClient
	iAmLeader          bool
	queue              *TimedQueue
	newLogEntryChannel chan *LogEntry
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
	}
	//edgeService.log.SetLogEntryUpdateChannel(edgeService.newLogEntryChannel)
	edgeService.queue = NewTimedQueue(configuration.Epoch.Duration, configuration.Epoch.MaxSize, edgeService.log.BatchedData)
	go edgeService.waitForLogEntryUpdate()
	return edgeService
}

func (e *EdgeService) Commit(ctx context.Context, commitData *CommitData) (*Dummy, error) {
	if e.iAmLeader {
		e.queue.Insert(commitData)
	} else {
		e.lc.sendCommitDataToLeader(commitData)
	}
	return &Dummy{}, nil
}

func (e *EdgeService) Read(ctx context.Context, val *KeyVal) (*ReadResponse, error) {
	key := string(val.Key)
	if v, ok := e.log.Read(key); ok {
		return &ReadResponse{
			Data: &KeyVal{
				Key:   val.Key,
				Value: v,
			}, Status: ResponseStatus_SUCCESS}, nil
	} else {
		return &ReadResponse{Data: &KeyVal{Key: val.Key, Value: []byte{0}}, Status: ResponseStatus_ERROR}, nil
	}
}

func (e *EdgeService) Propose(ctx context.Context, data *ProposeData) (d *Dummy, err error) {
	// TODO: validate header to ensure data was received from the right leader.
	if e.log.Propose(data.LogBlock.LogID, data.LogBlock.Data) {
		a := ConvertToAcceptMsg(e.log.logEntry[data.LogBlock.LogID], &e.config.Node, e.currentTerm, e.key)
		e.teClient.SendAccept(a)
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
		log.Printf("Verfied certificate received from TE")
		e.log.Certificate <- certificate
		err = nil
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
	if string(e.key.GetPublicKey()) == string(leaderConfig.LeaderPubKey.RawPublicKey) {
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
		log.Printf("Registered with TE. Registration configuration %v", regConfig)
		if regConfig.ClusterLeader != nil {
			e.leader = regConfig.ClusterLeader
			e.currentTerm = regConfig.ClusterLeader.TermID
			e.checkLeadershipStatusAndConnect(regConfig.ClusterLeader)
		} else {
			log.Printf("Leader does not exist. Persorming self promotion.")
			go e.teClient.SendSelfPromote(e.config.Node, e.currentTerm, e.key.GetPublicKey())
		}
		e.regConfig = regConfig
	} else {
		log.Printf("Error while registering with TE: %v", err)
	}
}

func (e *EdgeService) waitForLogEntryUpdate() {
	clusterClient := NewEdgeClient()
	for _, c := range e.config.ClusterNodes {
		addr := c.Ip + ":" + c.Port
		clusterClient.AddConnection(addr)
	}
	for l := range e.newLogEntryChannel {
		p := ConvertToProposeData(*l, &e.config.Node)
		clusterClient.SendProposal(p)
	}
}
