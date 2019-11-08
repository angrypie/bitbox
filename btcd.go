package bitbox

import (
	"errors"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/angrypie/procutil"
	"github.com/angrypie/rndport"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/google/uuid"
)

type Btcd struct {
	started     bool
	nodesNumber int
	nodes       []*btcdNode
}

func NewBtcd() *Btcd {
	return &Btcd{}
}

//Start runs specified number of bitcoind nodes in regtest mode.
func (b *Btcd) Start(nodes int) (err error) {
	if nodes < 1 {
		return errors.New("number of nodes should be greater than 0")
	}

	b.started = true
	b.nodesNumber = nodes

	b.nodes, err = newBtcdNodeSet(nodes)
	return
}

//Stop terminates all nodes, nnd cleans data directories.
func (b *Btcd) Stop() (err error) {
	//TODO Need to handle stop errors
	for _, node := range b.nodes {
		node.Stop()
	}
	return nil
}

//Info returns information about bitbox state.
func (b *Btcd) Info() (state *State) {
	var rpcPort string

	if len(b.nodes) > 0 {
		rpcPort = b.nodes[0].rpcport
	}

	return &State{
		Name:        "btcd",
		RPCPort:     rpcPort,
		IsStarted:   b.started,
		NodesNumber: b.nodesNumber,
		ZmqAddress:  "not supported",
	}
}

//Generate perform blocks mining.
func (b *Btcd) Generate(nodeIndex int, blocks uint32) (err error) {
	err = errors.New("TODO")
	return
}

//Send sends funds from node to specified address.
func (b *Btcd) Send(node int, address string, amount float64) (tx string, err error) {
	err = errors.New("TODO")
	return
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *Btcd) Balance(node int) (balance float64, err error) {
	err = errors.New("TODO")
	return
}

//Address generates new adderess of specified nodes wallet.
func (b *Btcd) Address(node int) (address string, err error) {
	err = errors.New("TODO")
	return
}

//GetRawTransaction returns raw transaction by hash.
func (b *Btcd) GetRawTransaction(txHash string) (result *btcutil.Tx, err error) {
	err = errors.New("TODO")
	return
}

//BlockHeight returns current block height.
func (b *Btcd) BlockHeight() (blocks int32, err error) {
	err = errors.New("TODO")
	return
}

//InitMempool makes mempool usable by sending transaction and generating blocks.
func (b *Btcd) InitMempool() (err error) {
	err = errors.New("TODO")
	return
}

type btcdNode struct {
	index   int
	datadir string
	port    string
	rpcport string
	client  *rpcclient.Client
	cmd     *exec.Cmd
}

func (bn *btcdNode) StartDaemon(masterRPCport string) (err error) {
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
		// First node will have empty masterRPCport but its' will not be executed
		opts = append(opts, "--connect=127.0.0.1:"+masterRPCport)
	} else {
		opts = append(opts,
			"--txindex",
		)
	}

	bn.cmd = exec.Command("btcd", opts...)
	err = bn.cmd.Start()
	if err != nil {
		log.Println("ERR starting btcd")
	}
	return
}

func (bn *btcdNode) Stop() {
	bn.client.Shutdown()
	err := procutil.Terminate(bn.cmd.Process)
	if err != nil {
		log.Println("ERR terminating btcd node process")
	}
	exec.Command("rm", "-rf", bn.datadir).Run()
}

func newBtcdNodeSet(number int) (nodes []*btcdNode, err error) {
	var masterRPCport string
	node, err := startBtcdNode(0, "")
	if err != nil {
		return
	}
	nodes = append(nodes, node)
	masterRPCport = node.port

	var wg sync.WaitGroup
	wg.Add(number - 1)
	for i := 1; i < int(number); i++ {
		i := i
		go func() {
			node, err = startBtcdNode(i, masterRPCport)
			if err != nil {
				return
			}
			nodes = append(nodes, node)
			wg.Done()
		}()
	}
	wg.Wait()
	return
}

func startBtcdNode(index int, masterRPCport string) (node *btcdNode, err error) {
	strIndex := uuid.New().String()
	datadir := "/tmp/bitbox_btcd_" + strIndex

	port, _ := rndport.GetAddress()
	rpcPort, _ := rndport.GetAddress()

	node = &btcdNode{
		index:   index,
		datadir: datadir,
		port:    port,
		rpcport: rpcPort,
	}

	//Create directory for node data
	err = exec.Command("mkdir", datadir).Run()
	if err != nil {
		log.Println("ERR creating datadir", err)
		return
	}

	client, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host: "127.0.0.1:" + node.rpcport, User: "test", Pass: "test",
			HTTPPostMode: true, DisableTLS: true,
		}, nil)
	if err != nil {
		log.Println("ERR rpc client not connected to node ", index, err)
		return
	}

	node.client = client
	_, err = node.client.GetInfo()
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		err = node.StartDaemon(masterRPCport)
		if err != nil {
			log.Println("ERR starting daemon", err)
			return
		}
		for {
			time.Sleep(time.Millisecond * 100)
			_, err = node.client.GetInfo()
			if err != nil {
				continue
			}
			break
		}
	}

	return
}
