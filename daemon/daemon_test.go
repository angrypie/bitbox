package main

import (
	"context"
	api "github.com/angrypie/bitbox/proto"
	"google.golang.org/grpc"
	"testing"
)

func TestDaemon(t *testing.T) {
	// Start daemon
	Start(10000)

	// Create grpc client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(":10000", opts...)
	if err != nil {
		t.Error("Cannt connect to daemon", err)
	}
	client := api.NewBitboxClient(conn)

	// Run 2 bitcoin nodes
	status, err := client.Start(context.Background(), &api.StartRequest{Nodes: 2})
	if err != nil {
		t.Error("Start rpc: ", err)
	}
	if status.GetStatus() != true {
		t.Error("Start Status should be true")
	}

	// Generate 102 blocks to earn btc for test
	status, err = client.Generate(context.Background(), &api.GenerateRequest{Node: 0, Blocks: 102})
	if err != nil {
		t.Error("Generate rpc: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Generate Status should be true")
	}

	// Generate addres for node1
	addrResp, err := client.Address(context.Background(), &api.NodeRequest{Node: 1})
	if err != nil {
		t.Error("Address rpc: ", err)
	}
	//TODO validate addres string
	if addrResp.GetAddress() == "" {
		t.Error("Wrong address ", addrResp.GetAddress())
	}
	// Send btc to node1
	sendToNode1Amount := 0.1
	status, err = client.Send(context.Background(), &api.SendRequest{
		Node: 0, Amount: sendToNode1Amount, Address: addrResp.GetAddress(),
	})

	if err != nil {
		t.Error("Send rpc: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Send Status should be true", status.Msg)
	}

	// Generate new block to confirm transaction
	client.Generate(context.Background(), &api.GenerateRequest{Node: 0, Blocks: 1})

	// Check if node1 receive btc
	balanceResp, err := client.Balance(context.Background(), &api.NodeRequest{Node: 1})

	if err != nil {
		t.Error("Balance rpc: ", err)
	}

	if balanceResp.GetBalance() != sendToNode1Amount {
		t.Error("Expected ", sendToNode1Amount, "on balance, got ", balanceResp.GetBalance())
	}

	// Stop bitcoind nodes
	status, err = client.Stop(context.Background(), &api.Empty{})
	if err != nil {
		t.Error("Status rpc: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Status should be true")
	}

}

func sleep() {
	<-make(chan bool)
}
