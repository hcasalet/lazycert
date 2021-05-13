package main

import (
	"flag"
	"github.com/hcasalet/lazycert/dump/lc"
	"google.golang.org/grpc"
	"log"
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
	id := flag.String("id", "1", "EdgeNode Identifier.")
	port := flag.String("port", "35001", "EdgeNode Identifier.")
	teAddr := flag.String("te", "localhost:35000", "EdgeNode Identifier.")
	flag.Parse()
	lis, err := net.Listen("tpc", "0.0.0.0:"+*port)
	if err != nil {
		log.Fatalf("Error starting the server at port 350000, %v", err)
	}

	s := grpc.NewServer()
	config := lc.NewConfig("E_" + *id)
	config.TEAddr = *teAddr
	config.Node.Port = *port
	edgeNode := lc.NewEdgeService(config)
	go edgeNode.RegisterWithTE()
	lc.RegisterEdgeNodeServer(s, edgeNode)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
