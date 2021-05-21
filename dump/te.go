package main

import (
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {

	log.Println("Trusted Entity Server")
	config := ReadTEArgs()
	lis, err := net.Listen("tcp", "0.0.0.0:"+config.Node.Port)
	if err != nil {
		log.Fatalf("Error starting the server at port 35000, %v", err)
	}
	s := grpc.NewServer()
	lc.RegisterTrustedEntityServer(s, lc.NewTrustedEntityService(config))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func ReadTEArgs() *lc.Config {
	ymlConfig := lc.NewYamlConfig()
	config := ymlConfig.GetTEConfig()
	return config
}
