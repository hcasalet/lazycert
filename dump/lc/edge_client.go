package lc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type EdgeClient struct {
	clientMap   map[string]*EdgeNodeClient
	connections map[string]*grpc.ClientConn
}

func NewEdgeClient() *EdgeClient {

	edgeNodeClients := &EdgeClient{
		clientMap:   make(map[string]*EdgeNodeClient),
		connections: make(map[string]*grpc.ClientConn),
	}

	return edgeNodeClients

}

func (e *EdgeClient) AddConnection(addr string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	edgeClient := NewEdgeNodeClient(conn)
	e.connections[addr] = conn
	e.clientMap[addr] = &edgeClient
}

func (e *EdgeClient) CloseAllConnections() {
	for k, v := range e.connections {
		err := v.Close()
		if err != nil {
			log.Printf("Could not close connection for node: %v", k)
		}
	}
}

func (e *EdgeClient) BroadcastLeaderStatus(leader LeaderConfig) {
	for addr, client := range e.clientMap {
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			_, err := (*client).LeaderStatus(ctx, &leader)
			if err != nil {
				log.Printf("Error occurred when sending leader status to '%v': %v", addr, err)
			}
			cancel()

		}
	}
}
