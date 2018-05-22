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
		opts = append(opts, "-connect=127.0.0.1:4900")
	} else {
		opts = append(opts,
			"-zmqpubhashtx=tcp://"+zmqaddress,
			"-zmqpubhashblock=tcp://"+zmqaddress,
			"-zmqpubrawblock=tcp://"+zmqaddress,
			"-zmqpubrawtx=tcp://"+zmqaddress,
			"-txindex=1",
		)
	}

	return exec.Command("bitcoind", opts...).Run()
}

func (bn *bitcoindNode) Stop() {
	bn.client.Shutdown()
	bn.Command("stop")
	exec.Command("rm", "-rf", bn.datadir).Run()
}

func (bn *bitcoindNode) Clean() {
	exec.Command("rm", "-r", bn.datadir).Run()
}

func (bn *bitcoindNode) Command(cmd ...string) error {
	args := append([]string{}, "-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test")
	full := append(args, cmd...)
	_, err := exec.Command("bitcoin-cli", full...).Output()
	return err
}

const MasterNodeRpcPort = "4901"

func startNode(index int) *bitcoindNode {
	//run bitcoin test box
	strIndex := strconv.Itoa(index)
	datadir := "/tmp/bitbox" + strIndex
	node := &bitcoindNode{
		index: index, datadir: datadir,
		port: "49" + strIndex + "0", rpcport: "49" + strIndex + "1",
	}

	exec.Command("mkdir", datadir).Run()

	err := node.Command("getnetworkinfo")
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		err := node.StartDaemon()
		if err != nil {
			log.Println(err)
			return nil
		}
		for {
			time.Sleep(time.Millisecond * 100)
			err := node.Command("getnetworkinfo")
			if err != nil {
				continue
			}
			break
		}
	}

	client, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host: "127.0.0.1:" + node.rpcport, User: "test", Pass: "test",
			HTTPPostMode: true, DisableTLS: true,
		}, nil)
	if err != nil {
		log.Println("Rpc client not connected to node ", index, err)
		return nil
	}
	node.client = client
	return node
}
