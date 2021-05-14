package main

import (
	"flag"
	"fmt"
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
	"math"
	"net"
)

/*func main() {
	//// Open the Badger database located in the /tmp/badger directory.
	//// It will be created if it doesn't exist.
	//db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//// Your code here…
	d := lc.Dummy{}
	fmt.Println(d)


}*/

func main() {
	config := ReadArgs()
	lis, err := net.Listen("tpc", "0.0.0.0:"+config.Node.Port)
	if err != nil {
		log.Fatalf("Error starting the server at port 350000, %v", err)
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
	flag.Parse()
	ymlConfig := NewLCConfig()
	port := ymlConfig.viper.GetString("edge_nodes." + *id + ".port")
	host := ymlConfig.viper.GetString("edge_nodes." + *id + ".host")
	tehost := ymlConfig.viper.GetString("te.host")
	teport := ymlConfig.viper.GetString("te.port")
	nodeCount := len(ymlConfig.viper.GetStringMap("edge_nodes"))
	config := lc.NewConfig("E_" + *id)
	config.TEAddr = fmt.Sprintf("%v:%v", tehost, teport)
	config.Node.Port = port
	config.Node.Ip = host
	config.Node.Uuid = *id
	config.F = int(math.Ceil(float64(nodeCount) / 2))
	return config
}
