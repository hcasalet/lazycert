package paxos

import (
	hm "github.com/cornelk/hashmap"
	"github.com/hcasalet/lazycert/dump/lc"
	"golang.org/x/net/context"
	"log"
)

type PaxosReplica struct {
	ballotNumber uint32
	dataStore    hm.HashMap
	id           string
	config       *lc.Config
}

func NewPaxosReplica(c *lc.Config) *PaxosReplica {
	pr := &PaxosReplica{
		ballotNumber: 0,
		dataStore:    hm.HashMap{},
		id:           c.Node.Uuid,
		config:       c,
	}
	return pr
}

func (p *PaxosReplica) Read(ctx context.Context, kv *KV) (*KV, error) {
	k := string(kv.Key)
	v, ok := p.dataStore.Get(k)
	resp := &KV{
		Key:   kv.Key,
		Value: nil,
	}
	if ok {
		resp.Value = v.([]byte)
	}
	return resp, nil
}

func (p *PaxosReplica) Write(ctx context.Context, kv *KV) (*KV, error) {
	qclients := p.createPXClient()
	p.ballotNumber += 1
	ballot := &Ballot{
		N: p.ballotNumber,
	}
	status, _ := qclients.SendProposal(ballot, p.config.F+1)
	if status {
		data := &Data{
			B:  ballot,
			Kv: kv,
		}
		qclients.SendAccept(data)
		p.updateDataStore(kv)
	}
	return kv, nil
}

func (p *PaxosReplica) createPXClient() *PXClient {
	qclients := NewPXClient()
	for _, c := range p.config.ClusterNodes {
		addr := c.Ip + ":" + c.Port
		qclients.AddConnection(addr)
	}
	return qclients
}

func (p *PaxosReplica) updateDataStore(kv *KV) {
	p.dataStore.Insert(string(kv.Key), kv.Value)
}

func (p *PaxosReplica) Prepare(ctx context.Context, ballot *Ballot) (*Promise, error) {
	log.Printf("Current ballot number: %v. Prepare received for ballot: %v\n", p.ballotNumber, ballot)
	promise := &Promise{
		Status: Status_FAIL,
		B: &Ballot{
			N: p.ballotNumber,
		},
	}
	if ballot.N > p.ballotNumber {
		promise.B = ballot
		promise.Status = Status_PASS
		p.ballotNumber = ballot.N
	}
	log.Printf("Returning promise: %v\n", promise)
	return promise, nil
}

func (p *PaxosReplica) Accept(ctx context.Context, data *Data) (*Dummy, error) {
	log.Printf("Ballot number: %v, Accept received for: %v", p.ballotNumber, data)
	if data.B.N == p.ballotNumber {
		p.updateDataStore(data.Kv)
		go p.sendLearn(data)
	}
	return &Dummy{}, nil
}

func (p *PaxosReplica) Learn(ctx context.Context, data *Data) (*Dummy, error) {
	p.updateDataStore(data.Kv)
	return &Dummy{}, nil
}

func (p *PaxosReplica) sendLearn(data *Data) {
	qclients := p.createPXClient()
	qclients.SendLearn(data)
}
