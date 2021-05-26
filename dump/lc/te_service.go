package lc

import (
	"golang.org/x/net/context"
	"log"
	"math/rand"
	"sort"
)

type TrustedEntityService struct {
	privateKey                *Key
	termID                    int32
	leaderConfig              LeaderConfig
	registeredNodes           map[string]*EdgeNodeConfig
	registrationConfiguration *RegistrationConfig
	selfPromotions            map[int32][]string
	configuration             *Config
	voteMap                   map[int32]map[string]Vote
	currentLogID              int32
	certifiedLogIDs           map[int32]bool
	certificates              map[int32]*Certificate
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
			ClusterLeader: &LeaderConfig{
				ID:           "",
				TermID:       0,
				Node:         nil,
				LeaderPubKey: nil,
			},
			LogPosition: 0,
		},
		configuration:   config,
		voteMap:         make(map[int32]map[string]Vote),
		currentLogID:    0,
		certifiedLogIDs: make(map[int32]bool),
		certificates:    make(map[int32]*Certificate),
	}
}

func (t *TrustedEntityService) Register(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*RegistrationConfig, error) {
	nodeID := getNodeID(ctx, edgeNodeConfig.Node)
	t.registeredNodes[nodeID] = edgeNodeConfig
	return t.registrationConfiguration, nil
}

func (t *TrustedEntityService) Accept(ctx context.Context, acc *AcceptMsg) (*Dummy, error) {
	nodeID := getNodeID(ctx, acc.Header.Node)
	log.Printf("Received accept message: %v,From %v\n", acc, nodeID)
	if acc.TermID == t.termID {
		logID := acc.Block.LogID
		if _, ok := t.voteMap[logID]; !ok {
			t.voteMap[logID] = make(map[string]Vote)
			t.certifiedLogIDs[logID] = false
		}
		if _, ok := t.voteMap[logID][nodeID]; !ok {
			if t.verifySignature(nodeID, acc.AcceptHash, acc.Signature) {
				t.voteMap[logID][nodeID] = Vote{
					Node:             acc.Header.Node,
					AcceptHash:       acc.AcceptHash,
					ReplicaSignature: acc.Signature,
				}
			}
		}
		// TODO Check if this is the right thing to do.
		if logID > t.currentLogID {
			t.currentLogID = logID
		}
	}
	go t.checkVotes()
	return &Dummy{}, nil
}

func (t *TrustedEntityService) GetCertificate(ctx context.Context, header *Header) (*Certificate, error) {
	panic("implement me")
}

func (t *TrustedEntityService) SelfPromotion(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*Dummy, error) {
	nodeID := getNodeID(ctx, edgeNodeConfig.Node)
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
		log.Printf("Edge node wants to initiate self promotion for the next term: %v", nextTerm)
		log.Printf("SelfPromotionList: %v", t.selfPromotions[nextTerm])
		if t.selfPromotions[nextTerm] == nil {
			t.selfPromotions[nextTerm] = make([]string, 1)
			t.selfPromotions[nextTerm][0] = nodeID
			log.Printf("SelfPromotionList: %v", t.selfPromotions[nextTerm])
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
	default:
		log.Printf("Cannot process self promotion for term %v. Proposed term should be %v.", edgeNodeConfig.TermID, nextTerm)
		// TODO if self promotion is for any other term other than nextTerm
		/**

		 */

	}
	go t.checkSelfPromotion()
	return &Dummy{}, nil
}

func getNodeID(ctx context.Context, n *NodeInfo) string {
	/*p, _ := peer.FromContext(ctx)
	log.Printf("Received registration request from %v", p)
	var srcIP string
	switch addr := p.Addr.(type) {
	case *net.TCPAddr:
		srcIP = addr.IP.String()
	}

	log.Printf("EdgeNode IP:Port %v:%v", srcIP, srcPort)*/
	nodeID := n.Ip + ":" + n.Port

	log.Printf("Node Identifier: %v", nodeID)
	return nodeID
}

func (t *TrustedEntityService) checkSelfPromotion() {
	nextTermID := t.termID + 1
	if votes, ok := t.selfPromotions[nextTermID]; ok && len(votes) >= (t.configuration.F+1) {
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
		for newLeaderIndex = leaderIndex; newLeaderIndex == leaderIndex; {
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
				Uuid: newLeader.Node.Uuid,
			},
			LeaderPubKey: &PublicKey{RawPublicKey: newLeader.PublicKey.RawPublicKey},
		}
		t.termID = nextTermID
		t.registrationConfiguration.ClusterLeader = &t.leaderConfig
		t.broadCastLeaderConfig()
	} else {
		log.Printf("Have not received enough votes to make leader selection.")
	}
}

func (t *TrustedEntityService) broadCastLeaderConfig() {
	edgeClient := t.getEdgeClientObject()
	edgeClient.BroadcastLeaderStatus(t.leaderConfig)
	edgeClient.CloseAllConnections()
}

func (t *TrustedEntityService) getEdgeClientObject() *EdgeClient {
	edgeClient := NewEdgeClient()
	for k := range t.registeredNodes {
		edgeClient.AddConnection(k)
	}
	return edgeClient
}

func (t *TrustedEntityService) verifySignature(nodeID string, messageHash []byte, signature []byte) (valid bool) {
	if edgeNodeConfig, ok := t.registeredNodes[nodeID]; ok {
		rawPublicKey := GetPublicKeyFromBytes(edgeNodeConfig.PublicKey.RawPublicKey)
		valid = VerifyMessage(messageHash, signature, rawPublicKey)
	}
	return valid
}

func (t *TrustedEntityService) checkVotes() {
	var sortedLogIDs []int
	for k := range t.voteMap {
		if done, ok := t.certifiedLogIDs[k]; ok && !done {
			sortedLogIDs = append(sortedLogIDs, int(k))
		}
	}
	log.Printf("Certified logIDs: %v", t.certifiedLogIDs)
	sort.Ints(sortedLogIDs)
	for _, logID := range sortedLogIDs {
		logIDint32 := int32(logID)
		if voteMap, ok := t.voteMap[logIDint32]; ok {
			count, acceptHash := t.countVotes(voteMap)
			if count >= t.configuration.F+1 {
				t.cerfityLogPosition(logIDint32, voteMap, acceptHash)
				t.certifiedLogIDs[logIDint32] = true
				log.Printf("Broadcasting certificates for logID: %v\n", logIDint32)
				t.broadCastCertificate(logIDint32)
			} else {
				break
			}
		}

	}
}

func (t *TrustedEntityService) cerfityLogPosition(logID int32, voteMap map[string]Vote, hash []byte) {
	if _, ok := t.certificates[logID]; !ok {
		teSignature := t.privateKey.Sign(hash)
		votes := make([]*Vote, len(voteMap))
		for _, v := range voteMap {
			votes = append(votes, &v)
		}

		t.certificates[logID] = &Certificate{
			LogID:       logID,
			AcceptHash:  hash,
			TeSignature: teSignature,
			Votes:       votes,
		}
	}
}

func (t *TrustedEntityService) countVotes(voteMap map[string]Vote) (int, []byte) {
	voteCount := make(map[string]int)
	count := 0
	var validAcceptHash []byte
	for _, vote := range voteMap {
		acceptHash := string(vote.AcceptHash)
		if _, ok := voteCount[acceptHash]; !ok {
			voteCount[acceptHash] = 0
		}
		voteCount[acceptHash] += 1
		if count < voteCount[acceptHash] {
			count = voteCount[acceptHash]
			validAcceptHash = vote.AcceptHash
		}
	}
	return count, validAcceptHash
}

func (t *TrustedEntityService) broadCastCertificate(logID int32) {
	edgeClient := t.getEdgeClientObject()
	edgeClient.BroadcastCertificate(t.certificates[logID])
	edgeClient.CloseAllConnections()
}
