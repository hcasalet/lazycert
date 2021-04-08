package lc

import (
	"golang.org/x/net/context"
)

type Server struct {

}

func (s Server) Commit(ctx context.Context, val *KeyVal) (*Dummy, error) {
	panic("implement me")
}

func (s Server) Read(ctx context.Context, val *KeyVal) (*KeyVal, error) {
	panic("implement me")
}

func (s Server) Propose(ctx context.Context, data *ProposeData) (*Dummy, error) {
	panic("implement me")
}

func (s Server) HeartBeat(ctx context.Context, info *HeartBeatInfo) (*Dummy, error) {
	panic("implement me")
}

func (s Server) Certification(ctx context.Context, certificate *Certificate) (*Dummy, error) {
	panic("implement me")
}

func (s Server) LeaderStatus(ctx context.Context, config *EdgeNodeConfig) (*Dummy, error) {
	panic("implement me")
}

