package lc

import "golang.org/x/net/context"

type TrustedEntityService struct {
	privateKey Key
}

func (t TrustedEntityService) Register(ctx context.Context, config *EdgeNodeConfig) (*RegistrationConfig, error) {
	panic("implement me")
}

func (t TrustedEntityService) Accept(ctx context.Context, ack *AcceptAck) (*Dummy, error) {
	panic("implement me")
}

func (t TrustedEntityService) GetCertificate(ctx context.Context, header *Header) (*Certificate, error) {
	panic("implement me")
}

func (t TrustedEntityService) SelfPromotion(ctx context.Context, config *EdgeNodeConfig) (*Dummy, error) {
	panic("implement me")
}
