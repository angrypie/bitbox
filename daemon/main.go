package main

import (
	api "github.com/angrypie/bitbox/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os/exec"
	"strconv"
	"time"
)

type Bitbox struct {
	started bool
	nodes   int32
}

func statusResp(status bool, msg string) (*api.StatusResponse, error) {
	return &api.StatusResponse{
		Status: status,
		Msg:    msg,
	}, nil
}

func (b *Bitbox) Start(ctx context.Context, req *api.StartRequest) (*api.StatusResponse, error) {
	nodes := req.GetNodes()
	if nodes < 1 {
		return statusResp(false, "Number of nodes should be greater than 0")
	}
	b.started = true
	b.nodes = nodes
	return nil, nil
}

func (b *Bitbox) Stop(ctx context.Context, req *api.Empty) (*api.StatusResponse, error) {
	return nil, nil
}

func (b *Bitbox) Generate(ctx context.Context, req *api.GenerateRequest) (*api.StatusResponse, error) {
	return nil, nil
}

func (b *Bitbox) Send(ctx context.Context, req *api.SendRequest) (*api.StatusResponse, error) {
	return nil, nil
}

func main() {
	server := grpc.NewServer()
	api.RegisterBitboxServer(server, &Bitbox{})
	n4 := startNode(4)
	defer n4.Stop()
}

func startNode(index int) *bitcoindNode {
	//run bitcoin test box
	datadir := strconv.Itoa(index)
	node := &bitcoindNode{
		index: index, datadir: datadir,
		port: "190" + datadir + "0", rpcport: "190" + datadir + "1",
	}

	exec.Command("mkdir", datadir).Run()

	err := node.Command("getinfo")
	if err != nil {
		node.Command("stop")
		time.Sleep(time.Second)
		node.Start()
		for {
			time.Sleep(time.Second)
			err := node.Command("getinfo")
			if err != nil {
				log.Println(err)
				continue
			}
			break
		}
		//TODO remove this line
		node.Command("generate", "101")
		time.Sleep(time.Second)
	}
	return node
}

type bitcoindNode struct {
	index   int
	datadir string
	port    string
	rpcport string
}

func (bn *bitcoindNode) Start() error {
	return exec.Command(
		"bitcoind", "-regtest", "-daemon",
		"-datadir="+bn.datadir, "-port="+bn.port,
		"-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test",
	).Run()
}

func (bn *bitcoindNode) Stop() {
	log.Println("STOP")
	defer bn.Command("stop")
	defer exec.Command("rm", "-r", bn.datadir).Run()
}

func (bn *bitcoindNode) Command(cmd ...string) error {
	args := append([]string{}, "-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test")
	full := append(args, cmd...)
	log.Println(full)
	return exec.Command(
		"bitcoin-cli", full...).Run()
}
