package main

import (
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {

	log.Println("Trusted Entity Server")

	lis, err := net.Listen("tpc", "35000")
	if err != nil {
		log.Fatalf("Error starting the server at port 350000, %v", err)
	}

	s := grpc.NewServer()
	lc.RegisterEdgeNodeServer(s, &lc.TrustedEntityService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
