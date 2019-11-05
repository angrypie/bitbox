package bitbox

import "github.com/btcsuite/btcutil"

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

//New creates new Bitbox client
func New() (bitbox Bitbox) {
	return &Bitcoind{}
}

//State represent current bitbox state, contain useful info.
type State struct {
	RPCPort     string
	ZmqAddress  string
	IsStarted   bool
	NodesNumber int
}
