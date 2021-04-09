package main

import (
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
	lis, err := net.Listen("tpc", "35000")
	if err != nil {
		log.Fatalf("Error starting the server at port 350000, %v", err)
	}

	s := grpc.NewServer()
	lc.RegisterEdgeNodeServer(s, &lc.EdgeService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
