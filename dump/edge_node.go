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
	flag.Parse()
	ymlConfig := lc.NewLCConfig()
	port := ymlConfig.Viper.GetString("edge_nodes." + *id + ".port")
	host := ymlConfig.Viper.GetString("edge_nodes." + *id + ".host")
	tehost := ymlConfig.Viper.GetString("te.host")
	teport := ymlConfig.Viper.GetString("te.port")
	epochDuration := ymlConfig.Viper.GetInt("epoch.duration")
	epochMaxSize := ymlConfig.Viper.GetInt("epoch.maxsize")
	nodeCount := len(ymlConfig.Viper.GetStringMap("edge_nodes"))
	config := lc.NewConfig("E_" + *id)
	config.TEAddr = fmt.Sprintf("%v:%v", tehost, teport)
	config.Node.Port = port
	config.Node.Ip = host
	config.Node.Uuid = *id
	config.F = int(math.Ceil(float64(nodeCount) / 2))
	config.Epoch.Duration = epochDuration
	config.Epoch.MaxSize = epochMaxSize
	return config
}
