package paxos

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type PXClient struct {
	clientMap   map[string]*PaxosClient
	connections map[string]*grpc.ClientConn
}

func NewPXClient() *PXClient {

	edgeNodeClients := &PXClient{
		clientMap:   make(map[string]*PaxosClient),
		connections: make(map[string]*grpc.ClientConn),
	}

	return edgeNodeClients

}

func (e *PXClient) AddConnection(addr string) {
	conn, edgeClient := CreateConnectionToEdgeNode(addr)
	e.connections[addr] = conn
	e.clientMap[addr] = &edgeClient
}

func CreateConnectionToEdgeNode(addr string) (*grpc.ClientConn, PaxosClient) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	edgeClient := NewPaxosClient(conn)
	return conn, edgeClient
}

func (e *PXClient) CloseAllConnections() {
	for k, v := range e.connections {
		err := v.Close()
		if err != nil {
			log.Printf("Could not close connection for node: %v", k)
		}
	}
}

func (e *PXClient) SendProposal(p *Ballot, votes int) (bool, uint32) {
	promiseCount := 0
	var highestBallotNumber uint32 = 0
	status := false
	log.Printf("Propose message: %v", p)
	for addr, client := range e.clientMap {
		log.Printf("Sending propose data to: %v", addr)
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			promise, err := (*client).Prepare(ctx, p)
			if err != nil {
				log.Printf("Error occurred when sending Propose data to '%v': %v", addr, err)
			} else {
				if promise.Status == Status_PASS {
					promiseCount += 1
				}
				if promise.Status == Status_FAIL {
					if highestBallotNumber < promise.B.N {
						highestBallotNumber = promise.B.N
					}
				}
			}
			cancel()
		}
	}
	if promiseCount >= votes {
		status = true
	}
	return status, highestBallotNumber
}

func (e *PXClient) SendAccept(data *Data) {

	for addr, client := range e.clientMap {
		log.Printf("Sending propose data to: %v", addr)
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			_, err := (*client).Accept(ctx, data)
			if err != nil {
				log.Printf("Error occurred when sending Propose data to '%v': %v", addr, err)
			}
			cancel()
		}
	}
}

func (e *PXClient) SendLearn(data *Data) {
	for addr, client := range e.clientMap {
		log.Printf("Sending propose data to: %v", addr)
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			_, err := (*client).Learn(ctx, data)
			if err != nil {
				log.Printf("Error occurred when sending Propose data to '%v': %v", addr, err)
			}
			cancel()
		}
	}
}
