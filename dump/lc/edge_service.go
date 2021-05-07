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

func (s EdgeService) Commit(ctx context.Context, val *KeyVal) (*Dummy, error) {
	panic("implement me")
}

func (s EdgeService) Read(ctx context.Context, val *KeyVal) (*KeyVal, error) {
	panic("implement me")
}

func (s EdgeService) Propose(ctx context.Context, data *ProposeData) (*Dummy, error) {
	panic("implement me")
}

func (s EdgeService) HeartBeat(ctx context.Context, info *HeartBeatInfo) (*Dummy, error) {
	panic("implement me")
}

func (s EdgeService) Certification(ctx context.Context, certificate *Certificate) (*Dummy, error) {
	panic("implement me")
}

func (s EdgeService) LeaderStatus(ctx context.Context, config *LeaderConfig) (*Dummy, error) {
	panic("implement me")
}
