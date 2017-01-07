package main

import (
	//"github.com/btcsuite/btcd/chaincfg"
	//"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcrpcclient"
	//"github.com/btcsuite/btcutil"
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
	return exec.Command(
		"bitcoind", "-regtest", "-daemon",
		"-datadir="+bn.datadir, "-port="+bn.port,
		"-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test",
	).Run()
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
				log.Println("Node ", index, ": getinfo:\n", err)
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
	//defer client.Shutdown()
	return node
}
