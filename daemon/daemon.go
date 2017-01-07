package main

import (
	api "github.com/angrypie/bitbox/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

type Bitbox struct {
	started     bool
	numberNodes int
	nodes       []*bitcoindNode
}

func main() {
	Start(10000)
	<-make(chan bool)
}

func Start(port int) *grpc.Server {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		log.Fatal("failed to listen: ", port)
	}
	server := grpc.NewServer()
	api.RegisterBitboxServer(server, &Bitbox{})
	go server.Serve(lis)
	return server
}
