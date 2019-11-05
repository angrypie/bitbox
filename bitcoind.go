package bitbox

import (
	"errors"
	"log"
	"os/exec"
	"time"

	"github.com/angrypie/rndport"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/google/uuid"
)

//Bitcoind represent set API interface to multiple bitcoin nodes,
//that are running in regtest mode.
type Bitcoind struct {
	started     bool
	nodesNumber int
	nodes       []*bitcoindNode
}

//Start runs specified number of bitcoind nodes in regtest mode.
func (b *Bitcoind) Start(nodes int) (err error) {
	if nodes < 1 {
		return errors.New("Number of nodes should be greater than 0")
	}

	b.started = true
	b.nodesNumber = nodes

	b.nodes = newBitcoindNodeSet(nodes)

	return nil
}

//Stop terminates all nodes, nnd cleans data directories.
func (b *Bitcoind) Stop() (err error) {
	//TODO Need to handle stop and clean errors
	for _, node := range b.nodes {
		node.Stop()
		node.Clean()
	}
	return nil
}

//Info returns information about bitbox state.
func (b *Bitcoind) Info() *State {
	var rpcPort, zmqAddress string

	if len(b.nodes) > 0 {
		rpcPort = b.nodes[0].rpcport
		zmqAddress = b.nodes[0].zmqaddress
	}

	return &State{
		RPCPort:     rpcPort,
		ZmqAddress:  zmqAddress,
		IsStarted:   b.started,
		NodesNumber: b.nodesNumber,
	}
}

//Generate perform blocks mining.
func (b *Bitcoind) Generate(nodeIndex int, blocks uint32) (err error) {
	node := b.nodes[nodeIndex]

	_, err = node.client.Generate(blocks)
	if err != nil {
		return err
	}

	return nil
}

//Send sends funds from node to specified address.
func (b *Bitcoind) Send(node int, address string, amount float64) (tx string, err error) {
	addr, err := btcutil.DecodeAddress(address, &chaincfg.RegressionNetParams)
	if err != nil {
		return "", errors.New("Wrong addres: " + err.Error())
	}

	btcAmount, err := btcutil.NewAmount(amount)
	if err != nil {
		return "", errors.New("Wrong amount: " + err.Error())
	}

	n := b.nodes[node]
	hash, err := n.client.SendToAddress(addr, btcAmount)
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *Bitcoind) Balance(node int) (balance float64, err error) {
	n := b.nodes[node]

	amount, err := n.client.GetBalance("*")
	if err != nil {
		return 0, err
	}

	return amount.ToBTC(), nil
}

//Address generates new adderess of specified nodes wallet.
func (b *Bitcoind) Address(node int) (address string, err error) {
	n := b.nodes[node]

	addr, err := n.client.GetNewAddress("")
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

//GetRawTransaction returns raw transaction by hash.
func (b *Bitcoind) GetRawTransaction(txHash string) (result *btcutil.Tx, err error) {
	n := b.nodes[0]

	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, err
	}

	transaction, err := n.client.GetRawTransaction(hash)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

//BlockHeight returns current block height.
func (b *Bitcoind) BlockHeight() (blocks int32, err error) {
	n := b.nodes[0]

	info, err := n.client.GetBlockChainInfo()
	if err != nil {
		return 0, err
	}

	return info.Blocks, nil
}

//InitMempool makes mempool usable by sending transaction and generating blocks.
func (b *Bitcoind) InitMempool() (err error) {
	err = b.Generate(0, 150)
	if err != nil {
		return err
	}
	addr, err := b.Address(0)
	if err != nil {
		return err
	}

	height, err := b.BlockHeight()
	if err != nil {
		return err
	}

	if height >= 200 {
		return
	}
	for i := 0; i < 50; i++ {
		_, err := b.Send(0, addr, 0.00001)
		if err != nil {
			log.Println(err)
		}
		err = b.Generate(0, 1)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

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
	opts := append([]string{}, "-regtest", "-daemon",
		"-deprecatedrpc=estimatefee,generate",
		"-deprecatedrpc=generate",
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
