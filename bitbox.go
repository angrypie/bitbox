package bitbox

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

//Bitbox represent API interface to multiple bitcoin nodes,
//that are running in regtest/simnet mode.
type Bitbox interface {
	//Start runs specified number of bitcoind nodes in regtest mode.
	Start(nodes int) (err error)
	//Stop terminates all nodes, nnd cleans data directories.
	Stop() (err error)
	//Info returns information about bitbox state.
	Info() *State
	//Generate perform blocks mining.
	Generate(nodeIndex int, blocks uint32) (err error)
	//Send sends funds from node to specified address.
	Send(node int, address string, amount float64) (tx string, err error)
	//Balance returns avaliable balance of specified nodes wallet.
	Balance(node int) (balance float64, err error)
	//Address generates new adderess of specified nodes wallet.
	Address(node int) (address string, err error)
	//GetRawTransaction returns raw transaction by hash.
	GetRawTransaction(txHash string) (result *btcutil.Tx, err error)
	//BlockHeight returns current block height.
	BlockHeight() (blocks int32, err error)
	//InitMempool makes mempool usable by sending transaction and generating blocks.
	InitMempool() (err error)
}

//New creates new Bitbox client. Pass 'btcd' as first argument to use such backend,
//by default 'bitcoind' is used.
func New(args ...string) (bitbox Bitbox) {
	if len(args) > 0 {
		switch args[0] {
		case "btcd":
			return newBtcd()
		}
	}
	return newBitcoind()
}

//State represent current bitbox state, contain useful info.
type State struct {
	Name        string
	NodePort    string
	RPCPort     string
	ZmqAddress  string
	IsStarted   bool
	NodesNumber int
}

type Node interface {
	Start() error
	Stop() error
	Client() *rpcclient.Client
	Info() *State
}

type BitboxDefaults struct {
	Nodes []Node
}

//Stop terminates all nodes, nnd cleans data directories.
func (b *BitboxDefaults) Stop() (err error) {
	//TODO Need to handle stop errors
	for _, node := range b.Nodes {
		node.Stop()
	}
	return nil
}

//Generate perform blocks mining.
func (b *BitboxDefaults) Generate(node int, blocks uint32) (err error) {
	_, err = b.Client(node).Generate(blocks)
	if err != nil {
		return err
	}

	return
}

//GetRawTransaction returns raw transaction by hash.
func (b *BitboxDefaults) GetRawTransaction(txHash string) (result *btcutil.Tx, err error) {
	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, err
	}

	transaction, err := b.Client(0).GetRawTransaction(hash)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

//BlockHeight returns current block height.
func (b *BitboxDefaults) BlockHeight() (blocks int32, err error) {

	info, err := b.Client(0).GetBlockChainInfo()
	if err != nil {
		return 0, err
	}

	return info.Blocks, nil
}

func (b *BitboxDefaults) Send(node int, address string, amount float64) (tx string, err error) {
	addr, err := btcutil.DecodeAddress(address, &chaincfg.RegressionNetParams)
	if err != nil {
		return "", errors.New("Wrong addres: " + err.Error())
	}

	btcAmount, err := btcutil.NewAmount(amount)
	if err != nil {
		return "", errors.New("Wrong amount: " + err.Error())
	}

	hash, err := b.Client(node).SendToAddress(addr, btcAmount)
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *BitboxDefaults) AccountBalance(node int, account string) (balance float64, err error) {
	amount, err := b.Client(node).GetBalance(account)
	if err != nil {
		return 0, err
	}

	return amount.ToBTC(), nil
}

//Address generates new adderess of specified nodes wallet.
func (b *BitboxDefaults) AccountAddress(node int, account string) (address string, err error) {
	addr, err := b.Client(node).GetNewAddress(account)
	if err != nil {
		return
	}

	return addr.String(), nil
}

func (b *BitboxDefaults) Client(nodeIndex int) *rpcclient.Client {
	if len(b.Nodes) <= nodeIndex {
		return nil
	}
	return b.Nodes[nodeIndex].Client()
}
