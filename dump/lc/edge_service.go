package lc

import (
	"golang.org/x/net/context"
	"log"
)

type EdgeService struct {
	dbDict      map[string][]byte
	logEntry    map[int32]*LogEntry
	key         *Key
	leader      *LeaderConfig
	currentTerm int32
	config      *Config
	teClient    *TEClient
	logPosition int32
	regConfig   *RegistrationConfig
}

func NewEdgeService(configuration *Config) *EdgeService {
	key := NewKey(configuration.PrivateKeyFileName)
	tec := NewTEClient(configuration.TEAddr)
	return &EdgeService{
		key:         key,
		dbDict:      make(map[string][]byte),
		logEntry:    make(map[int32]*LogEntry),
		leader:      nil,
		currentTerm: 0,
		config:      configuration,
		teClient:    tec,
		logPosition: 0,
	}
}

func (e *EdgeService) Commit(ctx context.Context, val *CommitData) (*Dummy, error) {
	panic("implement me")
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
	panic("implement me")
}

func (e *EdgeService) LeaderStatus(ctx context.Context, config *LeaderConfig) (*Dummy, error) {
	e.leader = config
	e.currentTerm = config.TermID

	return &Dummy{}, nil
}

func (e *EdgeService) RegisterWithTE() {
	regConfig, err := e.teClient.Register(PublicKey{
		RawPublicKey: e.key.GetPublicKey(),
	}, e.config.Node, e.currentTerm)
	if err == nil {
		e.leader = regConfig.ClusterLeader
		e.currentTerm = regConfig.ClusterLeader.TermID
		e.logPosition = regConfig.LogPosition
		e.regConfig = regConfig
	} else {
		log.Printf("Error while registering with TE: %v", err)
	}
}
