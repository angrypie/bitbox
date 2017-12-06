package bitbox

import (
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

type bitcoindNode struct {
	index   int
	datadir string
	port    string
	rpcport string
	client  *rpcclient.Client
}

func (bn *bitcoindNode) StartDaemon() error {
	//TODO
	zmqaddress := "127.0.0.1:28333"
	opts := append([]string{}, "-regtest", "-daemon",
		"-datadir="+bn.datadir, "-port="+bn.port,
		"-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test",
	)

	if bn.index > 0 {
		opts = append(opts, "-connect=127.0.0.1:19000")
	} else {
		opts = append(opts,
			"-zmqpubhashtx=tcp://"+zmqaddress,
			"-zmqpubhashblock=tcp://"+zmqaddress,
			"-zmqpubrawblock=tcp://"+zmqaddress,
			"-zmqpubrawtx=tcp://"+zmqaddress,
		)
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
		node.StartDaemon()
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

	client, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host: "localhost:" + node.rpcport, User: "test", Pass: "test",
			HTTPPostMode: true, DisableTLS: true,
		}, nil)
	if err != nil {
		log.Println("Rpc client not connected to node ", index)
	}
	node.client = client
	return node
}
