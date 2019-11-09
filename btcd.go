package bitbox

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/angrypie/procutil"
	"github.com/angrypie/rndport"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/google/uuid"
)

const defaultAccount = "default"

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
	node := b.nodes[nodeIndex]

	_, err = node.client.Generate(blocks)
	if err != nil {
		return err
	}

	return
}

//Send sends funds from node to specified address.
func (b *Btcd) Send(node int, address string, amount float64) (tx string, err error) {
	return b.send(node, defaultAccount, address, amount)
}

func (b *Btcd) send(node int, account, address string, amount float64) (tx string, err error) {
	addr, err := btcutil.DecodeAddress(address, &chaincfg.SimNetParams)
	if err != nil {
		return "", errors.New("Wrong addres: " + err.Error())
	}

	btcAmount, err := btcutil.NewAmount(amount)
	if err != nil {
		return "", errors.New("Wrong amount: " + err.Error())
	}

	n := b.nodes[node]
	hash, err := n.client.SendFrom(account, addr, btcAmount)
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func (b *Btcd) sendFromMiner(node int, address string, amount float64) (tx string, err error) {
	return b.send(node, "imported", address, amount)
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *Btcd) Balance(node int) (balance float64, err error) {
	n := b.nodes[node]

	amount, err := n.client.GetBalance(defaultAccount)
	if err != nil {
		return 0, err
	}

	return amount.ToBTC(), nil
}

//Address generates new adderess of specified nodes wallet.
func (b *Btcd) Address(node int) (address string, err error) {
	n := b.nodes[node]

	addr, err := n.client.GetNewAddress(defaultAccount)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

//GetRawTransaction returns raw transaction by hash.
func (b *Btcd) GetRawTransaction(txHash string) (result *btcutil.Tx, err error) {
	err = errors.New("TODO")
	return
}

//BlockHeight returns current block height.
func (b *Btcd) BlockHeight() (blocks int32, err error) {
	n := b.nodes[0]

	info, err := n.client.GetBlockChainInfo()
	if err != nil {
		return 0, err
	}

	return info.Blocks, nil
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
		if bal, err := b.nodes[0].client.GetBalance("imported"); err == nil && bal.ToBTC() != 0 {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}

	for i := 0; i < 50; i++ {
		_, err := b.sendFromMiner(0, addr, 10)
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
	index         int
	datadir       string
	port          string
	rpcport       string
	walletRPCPort string
	client        *rpcclient.Client
	btcdCmd       *exec.Cmd
	walletCmd     *exec.Cmd
	btcKey        *btcKey
}

func (bn *btcdNode) StartDaemon(masterRPCport string) (err error) {
	opts := append([]string{},
		"--simnet",
		"--notls",
		"--miningaddr="+bn.btcKey.Address().String(),
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

	return
}

func (bn *btcdNode) Stop() {
	bn.client.Shutdown()
	err := procutil.Terminate(bn.btcdCmd.Process)
	if err != nil {
		log.Println("ERR terminating btcd node process")
	}

	err = procutil.Terminate(bn.walletCmd.Process)
	if err != nil {
		log.Println("ERR terminating wallet process")
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
	walletPort, _ := rndport.GetAddress()

	node = &btcdNode{
		index:         index,
		datadir:       datadir,
		port:          port,
		rpcport:       rpcPort,
		walletRPCPort: walletPort,
		btcKey:        newBtcKey(),
	}

	//Create directory for node data
	err = exec.Command("mkdir", datadir).Run()
	if err != nil {
		log.Println("ERR creating datadir", err)
		return
	}

	client, err := newRpcClient("127.0.0.1:" + node.rpcport)
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

		for i := 0; ; i++ {
			time.Sleep(time.Millisecond * 100)
			_, err = node.client.GetInfo()
			if err != nil {
				if i == 40 {
					fmt.Println("trying to get info", err)
					i = 0
				}
				continue
			}
			break
		}
	}

	err = node.client.WalletPassphrase("password", 1e6)
	if err != nil {
		return
	}

	wif, err := node.btcKey.GetWIF()
	if err != nil {
		return
	}

	err = node.client.ImportPrivKey(wif)
	if err != nil {
		return
	}

	return
}

type btcKey struct {
	key *hdkeychain.ExtendedKey
	net *chaincfg.Params
}

func newBtcKey() *btcKey {
	net := &chaincfg.SimNetParams
	seed, _ := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	masterKey, _ := hdkeychain.NewMaster(seed, net)
	return &btcKey{key: masterKey, net: net}
}

func (key *btcKey) Address() btcutil.Address {
	addr, _ := key.key.Address(key.net)
	return addr
}

func (key *btcKey) GetWIF() (wif *btcutil.WIF, err error) {
	priv, _ := key.key.ECPrivKey()
	return btcutil.NewWIF(priv, key.net, true)
}

func newRpcClient(host string) (client *rpcclient.Client, err error) {
	client, err = rpcclient.New(
		&rpcclient.ConnConfig{
			Host:         host,
			User:         "test",
			Pass:         "test",
			HTTPPostMode: true,
			DisableTLS:   true,
		}, nil)
	return
}
