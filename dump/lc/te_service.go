package lc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	"log"
	"math/rand"
	"net"
)

type TrustedEntityService struct {
	privateKey                *Key
	termID                    int32
	leaderConfig              LeaderConfig
	registeredNodes           map[string]*EdgeNodeConfig
	registrationConfiguration *RegistrationConfig
	selfPromotions            map[int32][]string
	configuration             *Config
}

func NewTrustedEntityService(config *Config) *TrustedEntityService {
	privatekey := NewKey(config.PrivateKeyFileName)
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
		configuration: config,
	}
}

func (t *TrustedEntityService) Register(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*RegistrationConfig, error) {
	nodeID := getNodeID(ctx, edgeNodeConfig)
	t.registeredNodes[nodeID] = edgeNodeConfig
	return t.registrationConfiguration, nil
}

func (t *TrustedEntityService) Accept(ctx context.Context, ack *AcceptMsg) (*Dummy, error) {

	panic("implement me")
}

func (t *TrustedEntityService) GetCertificate(ctx context.Context, header *Header) (*Certificate, error) {
	panic("implement me")
}

func (t *TrustedEntityService) SelfPromotion(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*Dummy, error) {
	nodeID := getNodeID(ctx, edgeNodeConfig)
	log.Printf("Received self promotion from: %v", nodeID)
	nextTerm := t.termID + 1
	switch edgeNodeConfig.TermID {
	case t.termID:
		log.Println("Edge node term id = TE term ID. SP is for the current term.")
		/**
		If a leader has been assigned for this term, log the request, respond via leader status api to inform the node.
		If not, log the request.
		*/
	case nextTerm:
		log.Println("Edge node wants to initiate self promotion for the next term.")
		if t.selfPromotions[nextTerm] == nil {
			t.selfPromotions[nextTerm] = make([]string, 1)
			t.selfPromotions[t.termID][0] = nodeID
		} else {
			found := false
			for _, id := range t.selfPromotions[nextTerm] {
				if id == nodeID {
					found = true
					break
				}
			}
			if !found {
				t.selfPromotions[nextTerm] = append(t.selfPromotions[nextTerm], nodeID)
			} else {
				log.Printf("Received duplicate self promotion message from node: " + nodeID)
			}
		}
	case 0:
		log.Println("Initial self promotion. No leader identified.")

	}
	go t.checkSelfPromotion()
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

func (t *TrustedEntityService) checkSelfPromotion() {
	nextTermID := t.termID + 1
	if votes, ok := t.selfPromotions[nextTermID]; ok && len(votes) > (t.configuration.F+1) {
		log.Printf("Number of self promotions for termID %v is %v. Going ahead with picking a leader", nextTermID, len(votes))
		leaderIndex := -1
		for i, id := range votes {
			if id == t.leaderConfig.ID {
				log.Printf("Current leader will not be selected again. %v, %v", i, id)
				leaderIndex = i
				break
			}
		}
		var newLeaderIndex int
		for newLeaderIndex = leaderIndex; newLeaderIndex != leaderIndex; {
			newLeaderIndex = rand.Intn(len(votes))
		}
		newLeader := t.registeredNodes[votes[newLeaderIndex]]
		// TODO: Put old leader data into the database.
		t.leaderConfig = LeaderConfig{
			ID:     votes[newLeaderIndex],
			TermID: nextTermID,
			Node: &NodeInfo{
				Ip:   newLeader.Node.Ip,
				Port: newLeader.Node.Port,
			},
			LeaderPubKey: &PublicKey{RawPublicKey: newLeader.PublicKey.RawPublicKey},
		}
		t.termID = nextTermID
		t.broadCastLeaderConfig()
	} else {
		log.Printf("Have not received enough votes to make leader selection.")
	}
}

func (t *TrustedEntityService) broadCastLeaderConfig() {
	edgeClient := NewEdgeClient()
	for k, _ := range t.registeredNodes {
		edgeClient.AddConnection(k)
	}
	edgeClient.BroadcastLeaderStatus(t.leaderConfig)
	edgeClient.CloseAllConnections()
}
