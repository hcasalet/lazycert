package main

import (
	"github.com/hcasalet/lazycert/dump/lc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	log.Printf("This program generates synthetic transactions.")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:35005", opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	defer conn.Close()

	edgeNodeClient := lc.NewEdgeNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	//var data []*lc.KeyVal
	response, err := edgeNodeClient.Commit(ctx, &lc.CommitData{
		Data: []*lc.KeyVal{&lc.KeyVal{
			Key:   []byte("3"),
			Value: []byte("4"),
		}},
	})
	if err != nil {
		log.Fatalf("Error in sending txn to edge node.: %v", err)
	}
	log.Printf("Response from Edge node: %v", response)
}
