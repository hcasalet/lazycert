package lc

import (
	"golang.org/x/net/context"
)

type EdgeService struct {
	dbDict   map[string][]byte
	logEntry map[int32]*LogEntry
	key      *Key
}

func NewEdgeService(config *Config) *EdgeService {
	key := NewKey(config.PrivateKeyFileName)
	return &EdgeService{
		key:      key,
		dbDict:   make(map[string][]byte),
		logEntry: make(map[int32]*LogEntry),
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
	panic("implement me")
}
