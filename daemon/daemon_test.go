package main

import (
	"context"
	api "github.com/angrypie/bitbox/proto"
	"google.golang.org/grpc"
	"testing"
)

func TestDaemon(t *testing.T) {
	Start(10000)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(":10000", opts...)
	if err != nil {
		t.Error("Cannt connect to daemon", err)
	}

	client := api.NewBitboxClient(conn)

	status, err := client.Start(context.Background(), &api.StartRequest{Nodes: 2})
	if err != nil {
		t.Error("rpc Start: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Status should be true")
	}

	// Stop bitcoind nodes and remove folders
	status, err = client.Stop(context.Background(), &api.Empty{})
	if err != nil {
		t.Error("rpc Status: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Status should be true")
	}

}

func sleep() {
	<-make(chan bool)
}
