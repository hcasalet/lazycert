package lc

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type LeaderClient struct {
	addr   string
	conn   *grpc.ClientConn
	client *EdgeNodeClient
}

func NewLeaderClient() *LeaderClient {
	return &LeaderClient{
		addr: "",
		conn: nil,
	}
}

func (l *LeaderClient) ConnectToLeader(nodeInfo *NodeInfo) {
	leaderAddr := fmt.Sprintf("%v:%v", nodeInfo.Ip, nodeInfo.Port)
	if leaderAddr != l.addr {
		log.Printf("Old leader addr: %v, New leader addr: %v", l.addr, leaderAddr)
		if l.conn == nil {
			l.createNewConnection(leaderAddr)
		} else {
			l.CloseConnection()
			l.createNewConnection(leaderAddr)
		}
		l.addr = leaderAddr
	}
}

func (l *LeaderClient) CloseConnection() {
	err := l.conn.Close()
	if err != nil {
		log.Printf("Could not disconnect from leader: %v", l.addr)
	}
}

func (l *LeaderClient) createNewConnection(addr string) {
	conn, c := CreateConnectionToEdgeNode(addr)
	l.conn = conn
	l.client = &c
}

func (l *LeaderClient) sendCommitDataToLeader(data *CommitData) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := (*l.client).Commit(ctx, data)
	if err != nil {
		log.Printf("Error occurred while sending commit data to leader: %v, %v", l.addr, data)
	}

}
