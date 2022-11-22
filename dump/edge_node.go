package main

import (
	"flag"
	"github.com/hcasalet/lazycert/dump/lc"
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
	edgeNode := lc.NewEdgeService(config)
	go edgeNode.RegisterWithTE()
	lc.RegisterEdgeNodeServer(s, edgeNode)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func ReadArgs() *lc.Config {
	id := flag.String("id", "1", "EdgeNode Identifier.")
	ymlFile := flag.String("y", "./benchmarkconfig/configf1.yml", "Experiment configurations.")
	flag.Parse()
	log.Printf("Node UUID is %v", *id)
	ymlConfig := lc.NewYamlConfig(*ymlFile)
	config := ymlConfig.SetupEdgeConfig(id)
	return config
}
