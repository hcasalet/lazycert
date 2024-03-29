package main

import (
	"context"
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:35000", opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	defer conn.Close()

	teclient := lc.NewTrustedEntityClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	response, err := teclient.Register(ctx, &lc.EdgeNodeConfig{
		PublicKey: nil,
		Node: &lc.NodeInfo{
			Ip:   "localhost",
			Port: "35000",
		},
	})
	if err != nil {
		log.Fatalf("Error when registering with TE: %v", err)
	}
	log.Printf("Response from TE Server Registration: %v", response)
}
