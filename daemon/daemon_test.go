package main

import (
	api "github.com/angrypie/bitbox/proto"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestDaemon(t *testing.T) {
	Start(10000)
	time.Sleep(time.Second * 1)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(":10000", opts...)
	if err != nil {
		t.Error("Cannt connect to daemon", err)
	}

	api.NewBitboxClient(conn)
}
