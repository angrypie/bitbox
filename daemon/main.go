package main

import (
	api "github.com/angrypie/bitbox/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"
)

type Bitbox struct {
	started     bool
	numberNodes int
	nodes       []*bitcoindNode
}

func statusResp(status bool, msg string) (*api.StatusResponse, error) {
	return &api.StatusResponse{
		Status: status,
		Msg:    msg,
	}, nil
}

func (b *Bitbox) Start(ctx context.Context, req *api.StartRequest) (*api.StatusResponse, error) {
	//TODO chekc is it safe
	nodes := int(req.GetNodes())
	if nodes < 1 {
		return statusResp(false, "Number of nodes should be greater than 0")
	}
	b.started = true
	b.numberNodes = nodes

	for i := 0; i < int(nodes); i++ {
		node := startNode(i)
		b.nodes = append(b.nodes, node)
	}

	return statusResp(true, "Success")
}

func (b *Bitbox) Stop(ctx context.Context, req *api.Empty) (*api.StatusResponse, error) {
	//Stop all nodes
	for _, node := range b.nodes {
		node.Stop()
	}
	return statusResp(true, "Success")
}

func (b *Bitbox) Generate(ctx context.Context, req *api.GenerateRequest) (*api.StatusResponse, error) {
	nodeIndex := int(req.GetNode())
	blocks := int(req.GetBlocks())

	node := b.nodes[nodeIndex]
	if err := node.Command("generate", strconv.Itoa(blocks)); err != nil {
		return statusResp(false, err.Error())
	}

	return statusResp(true, "Success")
}

func (b *Bitbox) Send(ctx context.Context, req *api.SendRequest) (*api.StatusResponse, error) {
	nodeIndex := int(req.GetNode())
	amount := strconv.FormatFloat(req.GetAmount(), 'f', -1, 32)
	address := req.GetAddress()

	node := b.nodes[nodeIndex]
	if err := node.Command("sendtoaddress", address, amount); err != nil {
		return statusResp(false, err.Error())
	}

	return statusResp(true, "Success")
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
	bn.Command("stop")
	exec.Command("rm", "-r", bn.datadir).Run()
}

func (bn *bitcoindNode) Command(cmd ...string) error {
	args := append([]string{}, "-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test")
	full := append(args, cmd...)
	return exec.Command("bitcoin-cli", full...).Run()
}
