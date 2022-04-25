package main

import (
	"flag"
	"github.com/hcasalet/lazycert/dump/lc"
	"github.com/hcasalet/lazycert/paxos/paxos"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	config := ReadArgs()
	lis, err := net.Listen("tcp", "0.0.0.0:"+config.Node.Port)
	if err != nil {
		log.Fatalf("Error starting the server at port %v, %v", config.Node.Port, err)
	}
	s := grpc.NewServer()
	node := paxos.NewPaxosReplica(config)
	paxos.RegisterPaxosServer(s, node)
	paxos.RegisterReplicaServer(s, node)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func ReadArgs() *lc.Config {
	id := flag.String("id", "1", "Node Identifier.")
	flag.Parse()
	log.Printf("Node UUID is %v", *id)
	ymlConfig := lc.NewYamlConfig()
	config := ymlConfig.SetupEdgeConfig(id)
	return config
}
