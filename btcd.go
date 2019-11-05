package bitbox

import "github.com/btcsuite/btcutil"

type Btcd struct {
}

func NewBtcd() *Btcd {
	return &Btcd{}
}

//Start runs specified number of bitcoind nodes in regtest mode.
func (b *Btcd) Start(nodes int) (err error) {
	return
}

//Stop terminates all nodes, nnd cleans data directories.
func (b *Btcd) Stop() (err error) {
	return
}

//Info returns information about bitbox state.
func (b *Btcd) Info() (state *State) {
	return
}

//Generate perform blocks mining.
func (b *Btcd) Generate(nodeIndex int, blocks uint32) (err error) {
	return
}

//Send sends funds from node to specified address.
func (b *Btcd) Send(node int, address string, amount float64) (tx string, err error) {
	return
}

//Balance returns avaliable balance of specified nodes wallet.
func (b *Btcd) Balance(node int) (balance float64, err error) {
	return
}

//Address generates new adderess of specified nodes wallet.
func (b *Btcd) Address(node int) (address string, err error) {
	return
}

//GetRawTransaction returns raw transaction by hash.
func (b *Btcd) GetRawTransaction(txHash string) (result *btcutil.Tx, err error) {
	return
}

//BlockHeight returns current block height.
func (b *Btcd) BlockHeight() (blocks int32, err error) {
	return
}

//InitMempool makes mempool usable by sending transaction and generating blocks.
func (b *Btcd) InitMempool() (err error) {
	return
}
