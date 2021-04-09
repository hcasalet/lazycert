package lc

import (
	"golang.org/x/net/context"
)

type TrustedEntityService struct {
}

func (t TrustedEntityService) Commit(ctx context.Context, val *KeyVal) (*Dummy, error) {
	panic("implement me")
}

func (t TrustedEntityService) Read(ctx context.Context, val *KeyVal) (*KeyVal, error) {
	panic("implement me")
}

func (t TrustedEntityService) Propose(ctx context.Context, data *ProposeData) (*Dummy, error) {
	panic("implement me")
}

func (t TrustedEntityService) HeartBeat(ctx context.Context, info *HeartBeatInfo) (*Dummy, error) {
	panic("implement me")
}

func (t TrustedEntityService) Certification(ctx context.Context, certificate *Certificate) (*Dummy, error) {
	panic("implement me")
}

func (t TrustedEntityService) LeaderStatus(ctx context.Context, config *EdgeNodeConfig) (*Dummy, error) {
	panic("implement me")
}
