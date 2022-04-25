package main

import (
	"github.com/hcasalet/lazycert/paxos/paxos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	log.Printf("This program generates synthetic transactions.")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:35001", opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	defer conn.Close()

	edgeNodeClient := paxos.NewReplicaClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 100000*time.Millisecond)
	defer cancel()

	//var data []*lc.KeyVal
	for i := 0; i <= 100; i++ {

		starttime := time.Now()
		duration := time.Since(starttime)
		response, err := edgeNodeClient.Write(ctx, &paxos.KV{
			Key:   []byte("abcd" + string(i)),
			Value: []byte("abcd" + string(i+1)),
		})
		log.Printf("Response from node received in %v", duration.String())
		if err != nil {
			log.Fatalf("Error in sending write to node.: %v", err)
		}
		log.Printf("Write Response from node: %v", response)
		time.Sleep(time.Millisecond * 100)
	}
	log.Println("Read response from Edge node.")
	for i := 0; i <= 100; i++ {

		readResponse, _ := edgeNodeClient.Read(ctx, &paxos.KV{
			Key:   []byte("abcd" + string(i)),
			Value: nil,
		})
		log.Println("Read response for ", i, ": ", readResponse)
	}
}
