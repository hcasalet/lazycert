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
	selfPromotions            map[int32][]string
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
		selfPromotions:  make(map[int32][]string),
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
	nodeID := getNodeID(ctx, edgeNodeConfig)
	t.registeredNodes[nodeID] = edgeNodeConfig
	return t.registrationConfiguration, nil
}

func (t TrustedEntityService) Accept(ctx context.Context, ack *AcceptAck) (*Dummy, error) {
	panic("implement me")
}

func (t TrustedEntityService) GetCertificate(ctx context.Context, header *Header) (*Certificate, error) {
	panic("implement me")
}

func (t TrustedEntityService) SelfPromotion(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*Dummy, error) {
	nodeID := getNodeID(ctx, edgeNodeConfig)
	log.Printf("Received self promotion from: %v", nodeID)
	switch edgeNodeConfig.TermID {
	case t.termID:
		log.Println("Edge node term id = TE term ID. SP is for the current term.")
		/**
		If a leader has been assigned for this term, log the request, respond via leader status api to inform the node.
		If not, log the request.
		*/
	case t.termID + 1:
		log.Println("Edge node wants to initiate self promotion for the next term.")
		if t.selfPromotions[t.termID+1] == nil {
			t.selfPromotions[t.termID+1] = make([]string, 1)
			t.selfPromotions[t.termID][0] = nodeID
		} else {
			found := false
			for _, id := range t.selfPromotions[t.termID+1] {
				if id == nodeID {
					found = true
					break
				}
			}
			if !found {
				t.selfPromotions[t.termID] = append(t.selfPromotions[t.termID+1], nodeID)
			} else {
				log.Printf("Received duplicate self promotion message from node: " + nodeID)
			}
		}
	case 0:
		log.Println("Initial self promotion. No leader identified.")

	}
	return &Dummy{}, nil
}

func getNodeID(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) string {
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
	return nodeID
}
