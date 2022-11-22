package lc

import (
	"encoding/hex"
	hm "github.com/cornelk/hashmap"
	"golang.org/x/net/context"
	"log"
	"math/rand"
	"sort"
	"time"
)

type TrustedEntityService struct {
	privateKey                *Key
	termID                    int32
	leaderConfig              LeaderConfig
	registeredNodes           map[string]*EdgeNodeConfig
	registrationConfiguration *RegistrationConfig
	selfPromotions            map[int32][]string
	configuration             *Config
	//voteMap                   map[int32]map[string]Vote
	voteMap         hm.HashMap
	currentLogID    int32
	certifiedLogIDs hm.HashMap
	//certifiedLogIDs map[int32]bool
	//certificates              map[int32]*Certificate
	certificates              hm.HashMap
	votes                     hm.HashMap
	voteCount                 hm.HashMap
	maxVoted                  hm.HashMap
	logCertificationStartTime hm.HashMap
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
		voteMap:         hm.HashMap{},
		currentLogID:    0,
		certifiedLogIDs: hm.HashMap{},
		//certificates:              make(map[int32]*Certificate),
		certificates:              hm.HashMap{},
		votes:                     hm.HashMap{},
		voteCount:                 hm.HashMap{},
		maxVoted:                  hm.HashMap{},
		logCertificationStartTime: hm.HashMap{},
	}
}

func (t *TrustedEntityService) Register(ctx context.Context, edgeNodeConfig *EdgeNodeConfig) (*RegistrationConfig, error) {
	nodeID := getNodeID(ctx, edgeNodeConfig.Node)
	t.registeredNodes[nodeID] = edgeNodeConfig
	return t.registrationConfiguration, nil
}

func (t *TrustedEntityService) Accept(ctx context.Context, acc *AcceptMsg) (*Dummy, error) {
	nodeID := getNodeID(ctx, acc.Header.Node)
	log.Printf("Received accept message from node id %v.", nodeID)
	if acc.TermID == t.termID {
		if t.verifySignature(nodeID, acc.AcceptHash, acc.Signature) {
			t.addToVotesAndIncrementVoteCount(acc)
		}
	}
	go t.checkVotes2()
	return &Dummy{}, nil
}

func (t *TrustedEntityService) printVoteMap() {
	for v := range t.votes.Iter() {
		logID := v.Key.(int32)
		log.Printf("printVoteMap: LogID=%v", logID)
		voteMap := v.Value.(hm.HashMap)
		for m := range voteMap.Iter() {
			hashString := m.Key.(string)
			vote := m.Value.([]Vote)
			log.Printf("printVoteMap: Hash String:Vote = %v : %v", hashString, vote)
		}
	}
}

func (t *TrustedEntityService) addToVotesAndIncrementVoteCount(acc *AcceptMsg) {
	hashedString := hex.EncodeToString(acc.AcceptHash)
	log.Printf("Hash string = %v", hashedString)
	if _, ok := t.votes.Get(acc.Block.LogID); !ok {
		voteMap := hm.HashMap{}
		t.votes.Set(acc.Block.LogID, voteMap)
		t.voteCount.Set(acc.Block.LogID, 0)
		t.logCertificationStartTime.Set(acc.Block.LogID, time.Now())
	}
	v, _ := t.votes.Get(acc.Block.LogID)
	voteMap := v.(hm.HashMap)
	if _, ok := voteMap.Get(hashedString); !ok {
		//n := len(t.configuration.ClusterNodes)
		//l := make([]Vote, n, n)
		var l []Vote
		voteMap.Set(hashedString, l)
	}
	va, _ := voteMap.Get(hashedString)
	voteList := va.([]Vote)
	voteList = append(voteList, Vote{
		ReplicaSignature: acc.Signature,
	})
	voteMap.Set(hashedString, voteList)
	t.votes.Set(acc.Block.LogID, voteMap)
	max := 0
	maxVotedHash := hashedString
	for k := range voteMap.Iter() {
		hashString := k.Key.(string)
		voteCount := k.Value.([]Vote)
		if max < len(voteCount) {
			max = len(voteCount)
			maxVotedHash = hashString
		}
		log.Printf("VoteCount for message hash %v = %v. MAX = %v", hashString, len(voteCount), max)
	}
	//t.printVoteMap()
	t.voteCount.Set(acc.Block.LogID, max)
	t.maxVoted.Set(acc.Block.LogID, maxVotedHash)
}

func (t *TrustedEntityService) checkVotes2() {
	var sortedLogIDs []int
	for k := range t.voteCount.Iter() {
		logID := k.Key.(int32)
		if _, ok := t.certifiedLogIDs.Get(logID); !ok {
			sortedLogIDs = append(sortedLogIDs, int(logID))
		}
	}
	sort.Ints(sortedLogIDs)
	log.Printf("Log IDs Not certified yet: %v", sortedLogIDs)
	for _, v := range sortedLogIDs {
		logIDInt32 := int32(v)
		if vcount, ok := t.voteCount.Get(logIDInt32); ok {
			if count := vcount.(int); count > t.configuration.F {
				//Certify log ID.
				if logHash, ok := t.maxVoted.Get(logIDInt32); ok {
					t.certifiedLogIDs.Set(logIDInt32, logHash)
					certificate := t.createCertificate(logIDInt32, logHash.(string))
					t.certificates.Set(logIDInt32, certificate)
					if kt, ok := t.logCertificationStartTime.Get(logIDInt32); ok {
						startTime := kt.(time.Time)

						totalTime := time.Since(startTime)
						log.Printf("CERTIFICATION TIME FOR LOG ID: %v = %s", logIDInt32, totalTime)
					}
					t.broadCastCertificate(logIDInt32)
				} else {
					log.Printf("LogHash not found for log ID: %v", logIDInt32)
				}
			} else {
				log.Printf("VoteCount is not above F=%v threshold. LogID:VoteCount = %v:%v", t.configuration.F, logIDInt32, count)
			}
		} else {
			log.Printf("Could not find voteCount for logID: %v", logIDInt32)
		}
	}

}

func (t *TrustedEntityService) Accept2(ctx context.Context, acc *AcceptMsg) (*Dummy, error) {
	nodeID := getNodeID(ctx, acc.Header.Node)
	log.Printf("Received accept message: %v,From %v\n", acc, nodeID)
	if acc.TermID == t.termID {
		log.Println("TermID match: ", t.termID)
		logID := acc.Block.LogID
		log.Println("LogID: ", logID)
		if _, ok := t.voteMap.Get(logID); !ok {
			//t.voteMap[logID] = make(map[string]Vote)
			t.voteMap.Insert(logID, hm.HashMap{})
			t.certifiedLogIDs.Insert(logID, false)
		}

		v, ok := t.voteMap.Get(logID)
		if ok {

			voteMap := v.(hm.HashMap)
			_, ok := voteMap.Get(nodeID)
			log.Printf("logID: %v, nodeID %v, found: %v", logID, nodeID, ok)
			if !ok {
				if t.verifySignature(nodeID, acc.AcceptHash, acc.Signature) {
					log.Printf("Signature for nodeID %v, verified", nodeID)
					voteMap.Insert(nodeID, Vote{
						Node:             acc.Header.Node,
						AcceptHash:       acc.AcceptHash,
						ReplicaSignature: acc.Signature,
					})
					t.voteMap.Insert(logID, voteMap)
				} else {
					log.Printf("Signature NOT verified for node: %v", nodeID)
				}
			}
			log.Printf("LogID:VoteCount =  %v:%v ", logID, voteMap.Len())
		}
		//if _, ok := t.voteMap[logID][nodeID]; !ok {
		//	if t.verifySignature(nodeID, acc.AcceptHash, acc.Signature) {
		//		t.voteMap[logID][nodeID] = Vote{
		//			Node:             acc.Header.Node,
		//			AcceptHash:       acc.AcceptHash,
		//			ReplicaSignature: acc.Signature,
		//		}
		//	}
		//}
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
	log.Printf("Checking votes")
	var sortedLogIDs []int
	for k := range t.voteMap.Iter() {
		logID := k.Key.(int32)
		value, ok := t.certifiedLogIDs.Get(logID)

		if ok {
			exists := value.(bool)
			if exists {
				sortedLogIDs = append(sortedLogIDs, int(logID))
			}
		}
	}
	//for k := range t.voteMap {
	//	if done, ok := t.certifiedLogIDs[k]; ok && !done {
	//		sortedLogIDs = append(sortedLogIDs, int(k))
	//	}
	//}
	//log.Printf("Certified logIDs: %v", t.certifiedLogIDs)
	sort.Ints(sortedLogIDs)
	log.Printf("Sorted Log IDs: %v", sortedLogIDs)
	for _, logID := range sortedLogIDs {
		logIDint32 := int32(logID)
		if voteMap, ok := t.voteMap.Get(logIDint32); ok {
			vm := voteMap.(hm.HashMap)
			copiedVoteMap := t.copyVoteMap(vm)
			count, acceptHash := t.countVotes(copiedVoteMap)
			if count >= t.configuration.F+1 {
				t.cerfityLogPosition(logIDint32, copiedVoteMap, acceptHash)
				t.certifiedLogIDs.Insert(logIDint32, true)
				log.Printf("Broadcasting certificates for logID: %v\n", logIDint32)
				t.broadCastCertificate(logIDint32)
			} else {
				break
			}
		}

	}
}

func (t *TrustedEntityService) cerfityLogPosition(logID int32, voteMap map[string]Vote, hash []byte) {
	if _, ok := t.certificates.Get(logID); !ok {
		teSignature := t.privateKey.Sign(hash)
		votes := make([]*Vote, len(voteMap))
		for _, v := range voteMap {
			votes = append(votes, &v)
		}

		t.certificates.Set(logID, Certificate{
			LogID:       logID,
			AcceptHash:  hash,
			TeSignature: teSignature,
			Votes:       votes,
		})
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
	if c, ok := t.certificates.Get(logID); ok {
		edgeClient := t.getEdgeClientObject()
		certificate := c.(Certificate)
		edgeClient.BroadcastCertificate(&certificate)
		edgeClient.CloseAllConnections()
	}
}

func (t *TrustedEntityService) copyVoteMap(vm hm.HashMap) map[string]Vote {
	clonedMap := make(map[string]Vote)
	for k := range vm.Iter() {
		key := k.Key.(string)
		value := k.Value.(Vote)
		clonedMap[key] = value
	}
	return clonedMap
}

func (t *TrustedEntityService) createCertificate(idInt32 int32, hash string) Certificate {
	decodeString, err := hex.DecodeString(hash)
	c := Certificate{
		LogID:       idInt32,
		AcceptHash:  decodeString,
		TeSignature: nil,
		Votes:       nil,
	}
	if err == nil {
		signature := t.privateKey.Sign(decodeString)
		c.TeSignature = signature
	} else {
		log.Printf("Error decoding hash for log id %v.", idInt32)
	}
	return c
}
