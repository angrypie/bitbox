package main

import (
	"context"
	proto "github.com/angrypie/bitbox/proto"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestDaemon(t *testing.T) {
	// Start daemon
	Start(10000)
	delay := func() { time.Sleep(time.Millisecond * 100) }

	// Create grpc client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(":10000", opts...)
	if err != nil {
		t.Error("Cannt connect to daemon", err)
	}
	client := proto.NewBitboxClient(conn)

	// Run 2 bitcoin nodes
	status, err := client.Start(context.Background(), &proto.StartRequest{Nodes: 2})
	if err != nil {
		t.Error("Start rpc: ", err)
	}
	if status.GetStatus() != true {
		t.Error("Start Status should be true")
	}

	// Generate 102 blocks to earn btc for test
	status, err = client.Generate(context.Background(), &proto.GenerateRequest{Node: 0, Blocks: 102})
	if err != nil {
		t.Error("Generate rpc: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Generate Status should be true")
	}

	// Check if node1 receive btc
	beforeBalance, err := client.Balance(context.Background(), &proto.NodeRequest{Node: 1})
	if err != nil {
		t.Error("Balance rpc: ", err)
	}
	// Generate addres for node1
	addrResp, err := client.Address(context.Background(), &proto.NodeRequest{Node: 1})
	if err != nil {
		t.Error("Address rpc: ", err)
	}
	//TODO validate addres string
	if addrResp.GetAddress() == "" {
		t.Error("Wrong address ", addrResp.GetAddress())
	}

	// Send btc to node1
	var sendToNode1Amount float64 = 0.1
	status, err = client.Send(context.Background(), &proto.SendRequest{
		Node: 0, Amount: sendToNode1Amount, Address: addrResp.GetAddress(),
	})
	if err != nil {
		t.Error("Send rpc: ", err)
	}
	if status.GetStatus() != true {
		t.Error("Send Status should be true", status.Msg)
	}

	// Generate new block to confirm transaction
	client.Generate(context.Background(), &proto.GenerateRequest{Node: 0, Blocks: 1})
	delay()

	// Check if node1 receive btc
	for i, exit := 0, false; !exit; i++ {
		afterBalance, err := client.Balance(context.Background(), &proto.NodeRequest{Node: 1})
		if err != nil {
			t.Error("Balance rpc: ", err)
		}
		expectedBalance := float32(int32((beforeBalance.GetBalance()+sendToNode1Amount)*10000000) / 10000000)
		actualBalance := float32(int32(afterBalance.GetBalance()*10000000) / 10000000)
		if expectedBalance != actualBalance {
			if i > 20 {
				exit = true
				t.Error("Expceted balance ", expectedBalance, " got ", actualBalance)
			} else {
				delay()
			}
		} else {
			exit = true
		}
	}

	// Stop bitcoind nodes
	status, err = client.Stop(context.Background(), &proto.Empty{})
	if err != nil {
		t.Error("Status rpc: ", err)
	}

	if status.GetStatus() != true {
		t.Error("Status should be true")
	}

}
