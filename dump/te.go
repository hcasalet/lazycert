package main

import (
	"fmt"
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"math"
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
	tehost := ymlConfig.Viper.GetString("te.host")
	teport := ymlConfig.Viper.GetString("te.port")
	nodeCount := len(ymlConfig.Viper.GetStringMap("edge_nodes"))
	config := lc.NewConfig("TE")
	config.TEAddr = fmt.Sprintf("%v:%v", tehost, teport)
	config.Node.Port = teport
	config.Node.Ip = tehost
	config.Node.Uuid = "te"
	config.F = int(math.Ceil(float64(nodeCount) / 2))
	return config
}
