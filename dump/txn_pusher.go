package main

import (
	"flag"
	"fmt"
	"github.com/hcasalet/lazycert/dump/lc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {

	host := flag.String("h", "localhost", "HostName.")
	port := flag.String("p", "35003", "Port.")
	flag.Parse()
	log.Printf("This program generates synthetic transactions.")
	hostPort := fmt.Sprintf("%v:%v", *host, *port)
	log.Printf("Connecting to %v", hostPort)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(hostPort, opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	defer conn.Close()

	edgeNodeClient := lc.NewEdgeNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	//var data []*lc.KeyVal
	starttime := time.Now()
	response, err := edgeNodeClient.Commit(ctx, &lc.CommitData{
		Data: []*lc.KeyVal{&lc.KeyVal{
			Key:   []byte("3"),
			Value: []byte("8"),
		}},
	})
	duration := time.Since(starttime)
	log.Printf("Response from edge node received in %v", duration.String())
	if err != nil {
		log.Fatalf("Error in sending txn to edge node.: %v", err)
	}
	log.Printf("Response from Edge node: %v", response)
	log.Println("Read response from Edge node.")

	readResponse, err := edgeNodeClient.Read(ctx, &lc.KeyVal{Key: []byte("3")})
	log.Println(readResponse)
}
