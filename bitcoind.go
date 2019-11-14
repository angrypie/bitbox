package bitbox

import (
	"errors"
	"log"
	"os/exec"

	"github.com/angrypie/rndport"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/google/uuid"
)

type Bitcoind struct {
	BitboxDefaults
	started     bool
	nodesNumber int
}

func newBitcoind() *Bitcoind {
	return &Bitcoind{}
}

//Start runs specified number of bitcoind nodes in regtest mode.
func (b *Bitcoind) Start(nodes int) (err error) {
	if nodes < 1 {
		return errors.New("number of nodes should be greater than 0")
	}

	b.Nodes, err = newNodeSet(
		nodes,
		func(index int, masterNodePort string) (Node, error) {
			return startBitcoindNode(index, masterNodePort)
		},
	)
	if err != nil {
		return
	}

	b.started = true
	b.nodesNumber = nodes

	return
}

//Info returns information about bitbox state.
func (b *Bitcoind) Info() *State {
	var nodePort, zmqAddress, rpcPort string

	if len(b.Nodes) > 0 {
		info := b.Nodes[0].Info()
		nodePort = info.NodePort
		rpcPort = info.RPCPort
		zmqAddress = info.ZmqAddress
	}

	return &State{
		Name:        "bitcoind",
		NodePort:    nodePort,
		RPCPort:     rpcPort,
		ZmqAddress:  zmqAddress,
		IsStarted:   b.started,
		NodesNumber: b.nodesNumber,
	}
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *Bitcoind) Balance(node int) (balance float64, err error) {
	return b.BitboxDefaults.AccountBalance(node, "*")
}

//Address generates new adderess of specified nodes wallet.
func (b *Bitcoind) Address(node int) (address string, err error) {
	return b.BitboxDefaults.AccountAddress(node, "")
}

//InitMempool makes mempool usable by sending transaction and generating blocks.
func (b *Bitcoind) InitMempool() error {
	return initBitboxMempool(b)
}

type bitcoindNode struct {
	index          int
	datadir        string
	port           string
	rpcport        string
	client         *rpcclient.Client
	zmqaddress     string
	masterNodePort string
}

func (node *bitcoindNode) Client() *rpcclient.Client {
	return node.client
}

func (node *bitcoindNode) Info() *State {
	return &State{
		NodePort:   node.port,
		RPCPort:    node.rpcport,
		ZmqAddress: node.zmqaddress,
	}
}

func (node *bitcoindNode) Start() (err error) {
	zmqPort, err := rndport.GetAddress()
	if err != nil {
		return err
	}
	zmqaddress := "127.0.0.1:" + zmqPort
	node.zmqaddress = zmqaddress
	opts := append([]string{}, "-regtest", "-daemon",
		"-deprecatedrpc=estimatefee,generate",
		"-deprecatedrpc=generate",
		"-datadir="+node.datadir, "-port="+node.port,
		"-rpcport="+node.rpcport, "-rpcuser=test", "-rpcpassword=test",
	)

	if node.index > 0 {
		// First node will have empty masterNodeport but its' will not be executed
		opts = append(opts, "-connect=127.0.0.1:"+node.masterNodePort)
	} else {
		opts = append(opts,
			"-zmqpubhashtx=tcp://"+zmqaddress,
			"-zmqpubhashblock=tcp://"+zmqaddress,
			"-zmqpubrawblock=tcp://"+zmqaddress,
			"-zmqpubrawtx=tcp://"+zmqaddress,
			"-txindex=1",
		)
	}
	err = exec.Command("bitcoind", opts...).Run()
	if err != nil {
		return
	}

	node.client, err = newRpcClient("127.0.0.1:" + node.rpcport)
	if err != nil {
		log.Println("ERR rpc client not connected to node ", node.index, err)
		return
	}

	return waitUntilNodeStart(node)
}

func (node *bitcoindNode) Stop() (err error) {
	node.client.Shutdown()
	node.Command("stop")
	exec.Command("rm", "-rf", node.datadir).Run()
	return
}

func (node *bitcoindNode) Command(cmd ...string) error {
	args := append([]string{}, "-rpcport="+node.rpcport, "-rpcuser=test", "-rpcpassword=test")
	full := append(args, cmd...)
	_, err := exec.Command("bitcoin-cli", full...).Output()
	return err
}

func startBitcoindNode(index int, masterNodePort string) (node *bitcoindNode, err error) {
	//run bitcoin test box
	strIndex := uuid.New().String()
	datadir := "/tmp/bitbox_bitcoind_" + strIndex
	//TODO check errors
	port, _ := rndport.GetAddress()
	rpcPort, _ := rndport.GetAddress()
	node = &bitcoindNode{
		index:          index,
		datadir:        datadir,
		port:           port,
		rpcport:        rpcPort,
		masterNodePort: masterNodePort,
	}

	//Create directory for node data
	err = exec.Command("mkdir", datadir).Run()
	if err != nil {
		log.Println("ERR creating datadir", err)
		return
	}

	err = node.Start()
	if err != nil {
		return
	}

	return
}
