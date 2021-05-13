package main

import (
	"flag"
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {

	log.Println("Trusted Entity Server")
	port := flag.String("port", "35000", "EdgeNode Identifier.")
	flag.Parse()
	lis, err := net.Listen("tcp", "0.0.0.0:"+*port)
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
