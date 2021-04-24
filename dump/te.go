package main

import (
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {

	log.Println("Trusted Entity Server")

	lis, err := net.Listen("tcp", "localhost:35000")
	if err != nil {
		log.Fatalf("Error starting the server at port 35000, %v", err)
	}

	s := grpc.NewServer()
	config := lc.NewConfig("TE")

	lc.RegisterTrustedEntityServer(s, lc.NewTrustedEntityService(config))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
