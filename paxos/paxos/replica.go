package paxos

import (
	hm "github.com/cornelk/hashmap"
	"github.com/hcasalet/lazycert/dump/lc"
	"golang.org/x/net/context"
	"log"
	"time"
)

type PaxosReplica struct {
	ballotNumber    uint32
	dataStore       hm.HashMap
	id              string
	config          *lc.Config
	qclients        *PXClient
	writerChannel   chan *KV
	receiverChannel chan *KV
	batchProcessor  chan []*KV
}

func NewPaxosReplica(c *lc.Config) *PaxosReplica {
	pr := &PaxosReplica{
		ballotNumber:    0,
		dataStore:       hm.HashMap{},
		id:              c.Node.Uuid,
		config:          c,
		writerChannel:   make(chan *KV),
		receiverChannel: make(chan *KV),
		batchProcessor:  make(chan []*KV),
	}
	log.Println("Setting up connections to other replicas.")
	time.Sleep(time.Second * 10)
	pr.qclients = pr.createPXClient()
	log.Println("Connections setup.")
	//go pr.Writer()
	go pr.batchedWriter()
	go pr.processBatch()
	return pr
}

func (p *PaxosReplica) Read(ctx context.Context, kv *KV) (*KV, error) {
	//log.Printf("Read request received for key size: %v", len(kv.Key))
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
	//log.Printf("Write request received for key and value len: %v", len(kv.Key)+len(kv.Value))
	p.writerChannel <- kv
	//rkv := <-p.receiverChannel
	return kv, nil
}
func (p *PaxosReplica) batchedWriter() {
	maxSize := p.config.Epoch.MaxSize
	log.Printf("Max batch size is: %v", maxSize)
	batch := make([]*KV, maxSize)
	counter := 0
	for kv := range p.writerChannel {
		batch[counter] = kv
		counter += 1
		if counter >= maxSize {
			p.batchProcessor <- batch
			counter = 0
		}
	}
}

func (p *PaxosReplica) processBatch() {
	for kvs := range p.batchProcessor {
		p.ballotNumber += 1
		log.Printf("Processing batch for ballot number %v, batch size %v", p.ballotNumber, len(kvs))
		ballot := &Ballot{
			N: p.ballotNumber,
		}
		status, _ := p.qclients.SendProposal(ballot, p.config.F+1)
		if status {
			data := &Data{
				B:  ballot,
				Kv: kvs,
			}
			p.qclients.SendAccept(data)
			for _, kv := range kvs {
				p.updateDataStore(kv)
			}
		}
	}
}

func (p *PaxosReplica) Writer() {
	for kv := range p.writerChannel {

		p.ballotNumber += 1
		ballot := &Ballot{
			N: p.ballotNumber,
		}
		status, _ := p.qclients.SendProposal(ballot, p.config.F+1)
		if status {
			data := &Data{
				B:  ballot,
				Kv: []*KV{kv},
			}
			p.qclients.SendAccept(data)
			p.updateDataStore(kv)
		}
		p.receiverChannel <- kv
	}
}

func (p *PaxosReplica) createPXClient() *PXClient {
	pxClient := NewPXClient()
	for _, c := range p.config.ClusterNodes {
		addr := c.Ip + ":" + c.Port
		pxClient.AddConnection(addr)
	}
	return pxClient
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
		for _, kv := range data.Kv {
			p.updateDataStore(kv)
		}
		go p.sendLearn(data)
	}
	return &Dummy{}, nil
}

func (p *PaxosReplica) Learn(ctx context.Context, data *Data) (*Dummy, error) {
	for _, kv := range data.Kv {
		p.updateDataStore(kv)
	}
	return &Dummy{}, nil
}

func (p *PaxosReplica) sendLearn(data *Data) {
	p.qclients.SendLearn(data)
}
