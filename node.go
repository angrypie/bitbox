package bitbox

import (
	"log"
	"os/exec"
	"time"

	"github.com/angrypie/rndport"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/google/uuid"
)

type bitcoindNode struct {
	index      int
	datadir    string
	port       string
	rpcport    string
	client     *rpcclient.Client
	zmqaddress string
}

func newBitcoindNodeSet(number int) (nodes []*bitcoindNode) {
	var masterRPCport string
	for i := 0; i < int(number); i++ {
		node := startNode(i, masterRPCport)
		nodes = append(nodes, node)
		if i == 0 {
			masterRPCport = node.port
		}
	}
	return nodes
}

func (bn *bitcoindNode) StartDaemon(masterRPCport string) error {
	//TODO
	zmqPort, err := rndport.GetAddress()
	if err != nil {
		return err
	}
	zmqaddress := "127.0.0.1:" + zmqPort
	bn.zmqaddress = zmqaddress
	opts := append([]string{}, "-regtest", "-daemon", "-deprecatedrpc=estimatefee",
		"-datadir="+bn.datadir, "-port="+bn.port,
		"-rpcport="+bn.rpcport, "-rpcuser=test", "-rpcpassword=test",
	)

	if bn.index > 0 {
		// First node will have empty masterRPCport but its' will not be executed
		opts = append(opts, "-connect=127.0.0.1:"+masterRPCport)
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

func startNode(index int, masterRPCport string) *bitcoindNode {
	//run bitcoin test box
	strIndex := uuid.New().String()
	datadir := "/tmp/bitbox_" + strIndex
	//TODO check errors
	port, _ := rndport.GetAddress()
	rpcPort, _ := rndport.GetAddress()
	node := &bitcoindNode{
		index: index, datadir: datadir,
		port: port, rpcport: rpcPort,
	}

	exec.Command("mkdir", datadir).Run()

	err := node.Command("getnetworkinfo")
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		err := node.StartDaemon(masterRPCport)
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
