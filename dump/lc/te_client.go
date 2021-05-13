package lc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type TEClient struct {
	client TrustedEntityClient
}

func NewTEClient(addr string) *TEClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("Could not connect to TE server: %v", err)
	}
	teClient := &TEClient{}
	teClient.client = NewTrustedEntityClient(conn)
	return teClient
}

func (t *TEClient) Register(key PublicKey, node NodeInfo, termID int32) (*RegistrationConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	e := &EdgeNodeConfig{
		PublicKey: &key,
		Node:      &node,
		TermID:    termID,
	}
	reg, err := t.client.Register(ctx, e)
	return reg, err

}
