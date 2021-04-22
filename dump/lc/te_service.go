package lc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	"log"
)

type TrustedEntityService struct {
	privateKey                *Key
	termID                    int32
	leaderConfig              LeaderConfig
	registeredNodes           []*EdgeNodeConfig
	registrationConfiguration *RegistrationConfig
}

func NewTrustedEntityService() *TrustedEntityService {
	privatekey := NewKey()
	return &TrustedEntityService{
		privateKey: privatekey,
		termID:     0,
		leaderConfig: LeaderConfig{
			LeaderPubKey: nil,
			Node:         nil,
			TermID:       0,
		},
		registeredNodes: make([]*EdgeNodeConfig, 1),
		registrationConfiguration: &RegistrationConfig{
			TePublicKey: &PublicKey{
				RawPublicKey: privatekey.GetPublicKey(),
			},
			ClusterLeader: nil,
			LogPosition:   0,
		},
	}
}

func (t TrustedEntityService) Register(ctx context.Context, config *EdgeNodeConfig) (*RegistrationConfig, error) {
	peer, _ := peer.FromContext(ctx)
	log.Printf("Received registration request from %v", peer)
	log.Printf("Returning response to registration request: %v", t.registrationConfiguration)
	return t.registrationConfiguration, nil
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
