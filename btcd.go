package bitbox

import (
	"errors"
	"log"
	"os/exec"
	"time"

	"github.com/angrypie/procutil"
	"github.com/angrypie/rndport"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/google/uuid"
)

const defaultAccount = "default"

type Btcd struct {
	BitboxDefaults
	started     bool
	nodesNumber int
}

func newBtcd() *Btcd {
	return &Btcd{}
}

//Start runs specified number of bitcoind nodes in regtest mode.
func (b *Btcd) Start(nodes int) (err error) {
	if nodes < 1 {
		return errors.New("number of nodes should be greater than 0")
	}

	b.Nodes, err = newBtcdNodeSet(nodes)
	if err != nil {
		return
	}

	b.started = true
	b.nodesNumber = nodes

	return
}

//Info returns information about bitbox state.
func (b *Btcd) Info() (state *State) {
	var nodePort, rpcPort string

	if len(b.Nodes) > 0 {
		info := b.Nodes[0].Info()
		nodePort = info.NodePort
		rpcPort = info.RPCPort
	}

	return &State{
		Name:        "btcd",
		NodePort:    nodePort,
		RPCPort:     rpcPort,
		IsStarted:   b.started,
		NodesNumber: b.nodesNumber,
		ZmqAddress:  "not supported",
	}
}

//Send sends funds from node to specified address.
func (b *Btcd) Send(node int, address string, amount float64) (tx string, err error) {
	addr, err := btcutil.DecodeAddress(address, &chaincfg.SimNetParams)
	if err != nil {
		return
	}

	btcAmount, err := btcutil.NewAmount(amount)
	if err != nil {
		return
	}

	hash, err := b.Client(node).SendFrom(defaultAccount, addr, btcAmount)
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *Btcd) Balance(node int) (balance float64, err error) {
	return b.BitboxDefaults.AccountBalance(node, defaultAccount)
}

//Address generates new adderess of specified nodes wallet.
func (b *Btcd) Address(node int) (address string, err error) {
	return b.BitboxDefaults.AccountAddress(node, defaultAccount)
}

//InitMempool makes mempool usable by sending transaction and generating blocks.
func (b *Btcd) InitMempool() (err error) {
	err = b.Generate(0, 200)
	if err != nil {
		return err
	}

	addr, err := b.Address(0)
	if err != nil {
		return err
	}

	//TODO max waiting time
	for {
		if bal, err := b.Balance(0); err == nil && bal > 101 {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}

	for i := 0; i < 50; i++ {
		_, err := b.Send(0, addr, 2)
		if err != nil {
			log.Println(err)
		}
		err = b.Generate(0, 1)
		if err != nil {
			log.Println(err)
		}
	}
	return
}

type btcdNode struct {
	index          int
	datadir        string
	port           string
	rpcport        string
	walletRPCPort  string
	masterNodePort string
	client         *rpcclient.Client
	btcdCmd        *exec.Cmd
	walletCmd      *exec.Cmd
}

func (bn *btcdNode) Client() *rpcclient.Client {
	return bn.client
}

func (bn *btcdNode) Info() *State {
	return &State{
		NodePort: bn.port,
		RPCPort:  bn.rpcport,
	}
}

func (bn *btcdNode) Start() (err error) {
	opts := append([]string{},
		"--simnet",
		"--notls",
		"--datadir="+bn.datadir,
		"--listen=127.0.0.1:"+bn.port,
		"--rpclisten=127.0.0.1:"+bn.rpcport,
		"--rpcuser=test",
		"--rpcpass=test",
	)

	if bn.index > 0 {
		// First node will have empty masterNodePort but its' will not be executed
		opts = append(opts, "--connect=127.0.0.1:"+bn.masterNodePort)
	} else {
		opts = append(opts,
			"--txindex",
		)
	}
	bn.btcdCmd = exec.Command("btcd", opts...)
	err = bn.btcdCmd.Start()
	if err != nil {
		log.Println("ERR starting btcd")
		return
	}

	walletOpts := append([]string{},
		"--simnet",
		"--createtemp",
		"--noservertls",
		"--noclienttls",
		"--appdata="+bn.datadir,
		"--rpcconnect=127.0.0.1:"+bn.rpcport,
		"--rpclisten=127.0.0.1:"+bn.walletRPCPort,
		"--username=test",
		"--password=test",
	)

	bn.walletCmd = exec.Command("btcwallet", walletOpts...)
	err = bn.walletCmd.Start()
	if err != nil {
		log.Println("ERR starting btcd")
		return
	}

	bn.client, err = newRpcClient("127.0.0.1:" + bn.walletRPCPort)
	if err != nil {
		log.Println("ERR starting starting wallet")
		return
	}
	waitUntilNodeStart(bn)

	addr, err := bn.Client().GetNewAddress(defaultAccount)
	if err != nil {
		log.Println("ERR starting node: getting address")
		return
	}

	opts = append(opts, "--miningaddr="+addr.String())

	err = procutil.Terminate(bn.btcdCmd.Process)
	if err != nil {
		log.Println("ERR starting node: terminating node")
		return
	}

	bn.btcdCmd = exec.Command("btcd", opts...)
	err = bn.btcdCmd.Start()
	if err != nil {
		log.Println("ERR starting btcd second time")
		return
	}

	waitUntilNodeStart(bn)

	return
}

func (bn *btcdNode) Stop() (err error) {
	bn.client.Shutdown()
	err = procutil.Terminate(bn.btcdCmd.Process)
	if err != nil {
		log.Println("ERR terminating btcd node process")
	}

	err = procutil.Terminate(bn.walletCmd.Process)
	if err != nil {
		log.Println("ERR terminating wallet process")
	}
	exec.Command("rm", "-rf", bn.datadir).Run()
	return
}

func newBtcdNodeSet(number int) (nodes []Node, err error) {
	return newNodeSet(
		number,
		func(index int, masterNodePort string) (Node, error) {
			return startBtcdNode(index, masterNodePort)
		},
	)
}

func startBtcdNode(index int, masterNodePort string) (node *btcdNode, err error) {
	strIndex := uuid.New().String()
	datadir := "/tmp/bitbox_btcd_" + strIndex

	port, _ := rndport.GetAddress()
	rpcPort, _ := rndport.GetAddress()
	walletPort, _ := rndport.GetAddress()

	node = &btcdNode{
		index:          index,
		datadir:        datadir,
		port:           port,
		rpcport:        rpcPort,
		walletRPCPort:  walletPort,
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

	err = node.client.WalletPassphrase("password", 1e6)
	if err != nil {
		return
	}

	return
}
