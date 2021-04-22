package lc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	"log"
	"net"
)

type TrustedEntityService struct {
	privateKey                *Key
	termID                    int32
	leaderConfig              LeaderConfig
	registeredNodes           map[string]*EdgeNodeConfig
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
		registeredNodes: make(map[string]*EdgeNodeConfig),
		registrationConfiguration: &RegistrationConfig{
			TePublicKey: &PublicKey{
				RawPublicKey: privatekey.GetPublicKey(),
			},
			ClusterLeader: nil,
			LogPosition:   0,
		},
	}
}

func (t TrustedEntityService) Register(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*RegistrationConfig, error) {
	peer, _ := peer.FromContext(ctx)
	log.Printf("Received registration request from %v", peer)
	var srcIP string
	switch addr := peer.Addr.(type) {
	case *net.TCPAddr:
		srcIP = addr.IP.String()
	}
	srcPort := edgeNodeConfig.Node.Port
	log.Printf("EdgeNode IP:Port %v:%v", srcIP, srcPort)
	nodeID := srcIP + ":" + srcPort

	log.Printf("Node Identifier: %v", nodeID)
	t.registeredNodes[nodeID] = edgeNodeConfig
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
