package main

import (
	"github.com/btcsuite/btcrpcclient"
	"log"
	"os/exec"
	"strconv"
	"time"
)

type bitcoindNode struct {
	index   int
	datadir string
	port    string
	rpcport string
	client  *btcrpcclient.Client
}

func (bn *bitcoindNode) Start() error {
	opts := append([]string{}, "-regtest", "-daemon",
		"-datadir="+bn.datadir, "-port="+bn.port,
		"-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test",
	)

	if bn.index > 0 {
		opts = append(opts, "-connect=127.0.0.1:19000")
	}

	return exec.Command("bitcoind", opts...).Run()
}

func (bn *bitcoindNode) Stop() {
	bn.client.Shutdown()
	bn.Command("stop")
}

func (bn *bitcoindNode) Clean() {
	exec.Command("rm", "-r", bn.datadir).Run()
}

func (bn *bitcoindNode) Command(cmd ...string) error {
	args := append([]string{}, "-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test")
	full := append(args, cmd...)
	return exec.Command("bitcoin-cli", full...).Run()
}

func startNode(index int) *bitcoindNode {
	//run bitcoin test box
	strIndex := strconv.Itoa(index)
	datadir := "bitbox" + strIndex
	//TODO: dynamic port sesection
	node := &bitcoindNode{
		index: index, datadir: datadir,
		port: "190" + strIndex + "0", rpcport: "190" + strIndex + "1",
	}

	exec.Command("mkdir", datadir).Run()

	err := node.Command("getinfo")
	if err != nil {
		node.Command("stop")
		time.Sleep(time.Millisecond * 100)
		node.Start()
		for {
			time.Sleep(time.Millisecond * 100)
			err := node.Command("getinfo")
			if err != nil {
				log.Println(err)
				continue
			}
			break
		}
	}

	client, err := btcrpcclient.New(
		&btcrpcclient.ConnConfig{
			Host: "localhost:" + node.rpcport, User: "test", Pass: "test",
			HTTPPostMode: true, DisableTLS: true,
		}, nil)
	if err != nil {
		log.Println("Rpc client not connected to node ", index)
	}
	node.client = client
	return node
}
