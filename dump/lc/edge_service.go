package lc

import (
	"golang.org/x/net/context"
	"log"
)

type EdgeService struct {
	dbDict      map[string][]byte
	key         *Key
	log         *Log
	leader      *LeaderConfig
	currentTerm int32
	config      *Config
	teClient    *TEClient
	regConfig   *RegistrationConfig
	lc          *LeaderClient
	iAmLeader   bool
	queue       *TimedQueue
}

func NewEdgeService(configuration *Config) *EdgeService {
	key := NewKey(configuration.PrivateKeyFileName)
	tec := NewTEClient(configuration.TEAddr)
	edgeService := &EdgeService{
		key:         key,
		dbDict:      make(map[string][]byte),
		leader:      nil,
		currentTerm: 0,
		config:      configuration,
		teClient:    tec,
		lc:          nil,
		iAmLeader:   false,
		log:         NewLog(),
	}
	edgeService.queue = NewTimedQueue(configuration.Epoch.Duration, configuration.Epoch.MaxSize, edgeService.log.BatchedData)
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
	if v, ok := e.dbDict[key]; ok {
		return &ReadResponse{
			Data: &KeyVal{
				Key:   val.Key,
				Value: v,
			}, Status: ResponseStatus_SUCCESS}, nil
	} else {
		return &ReadResponse{Data: &KeyVal{Key: val.Key, Value: []byte{0}}, Status: ResponseStatus_ERROR}, nil
	}
}

func (e *EdgeService) Propose(ctx context.Context, data *ProposeData) (*Dummy, error) {
	panic("implement me")
}

func (e *EdgeService) HeartBeat(ctx context.Context, info *HeartBeatInfo) (*Dummy, error) {
	panic("implement me")
}

func (e *EdgeService) Certification(ctx context.Context, certificate *Certificate) (*Dummy, error) {
	e.log.Certificate <- certificate
	return &Dummy{}, nil
}

func (e *EdgeService) LeaderStatus(ctx context.Context, leaderConfig *LeaderConfig) (*Dummy, error) {
	e.leader = leaderConfig
	e.currentTerm = leaderConfig.TermID
	if string(e.key.GetPublicKey()) == string(leaderConfig.LeaderPubKey.RawPublicKey) {
		e.iAmLeader = true
	}
	if !e.iAmLeader {
		if e.lc == nil {
			e.lc = NewLeaderClient()
		}
		go e.lc.ConnectToLeader(e.leader.Node)
	}
	return &Dummy{}, nil
}

func (e *EdgeService) RegisterWithTE() {
	regConfig, err := e.teClient.Register(PublicKey{
		RawPublicKey: e.key.GetPublicKey(),
	}, e.config.Node, e.currentTerm)
	if err == nil {
		log.Printf("Registered with TE. Registration configuration %v", regConfig)
		e.leader = regConfig.ClusterLeader
		e.currentTerm = regConfig.ClusterLeader.TermID
		e.regConfig = regConfig
		e.lc = NewLeaderClient()
		e.lc.ConnectToLeader(e.leader.Node)
	} else {
		log.Printf("Error while registering with TE: %v", err)
	}
}
