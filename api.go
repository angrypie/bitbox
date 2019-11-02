package bitbox

import (
	"errors"
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

//Start runs specified number of bitcoind nodes in regtest mode.
func (b *Bitbox) Start(nodes int) (err error) {
	if nodes < 1 {
		return errors.New("Number of nodes should be greater than 0")
	}

	b.started = true
	b.numberNodes = nodes

	b.nodes = newBitcoindNodeSet(nodes)

	return nil
}

//Stop terminates all nodes, nnd cleans data directories.
func (b *Bitbox) Stop() (err error) {
	//TODO Need to handle stop and clean errors
	for _, node := range b.nodes {
		node.Stop()
		node.Clean()
	}
	return nil
}

//State represent current bitbox state, contain useful info.
type State struct {
	RPCPort    string
	ZmqAddress string
	IsStarted  bool
}

//Info returns information about bitbox state.
func (b *Bitbox) Info() *State {
	return &State{
		RPCPort:    b.nodes[0].rpcport,
		ZmqAddress: b.nodes[0].zmqaddress,
		IsStarted:  b.started,
	}
}

//Generate perform blocks mining.
func (b *Bitbox) Generate(nodeIndex int, blocks uint32) (err error) {
	node := b.nodes[nodeIndex]

	_, err = node.client.Generate(blocks)
	if err != nil {
		return err
	}

	return nil
}

//Send sends funds from node to specified address.
func (b *Bitbox) Send(node int, address string, amount float64) (tx string, err error) {
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
func (b *Bitbox) Balance(node int) (balance float64, err error) {
	n := b.nodes[node]

	amount, err := n.client.GetBalance("")
	if err != nil {
		return 0, err
	}

	return amount.ToBTC(), nil
}

//Address generates new adderess of specified nodes wallet.
func (b *Bitbox) Address(node int) (address string, err error) {
	n := b.nodes[node]

	addr, err := n.client.GetNewAddress("")
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

//GetRawTransaction returns raw transaction by hash.
func (b *Bitbox) GetRawTransaction(txHash string) (result *btcutil.Tx, err error) {
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
func (b *Bitbox) BlockHeight() (blocks int32, err error) {
	n := b.nodes[0]

	info, err := n.client.GetBlockChainInfo()
	if err != nil {
		return 0, err
	}

	return info.Blocks, nil
}

//EstimateFee returns estimated fee rate for specified number of blocks.
func (b *Bitbox) EstimateFee(numBlocks int64) (fee float64, err error) {
	n := b.nodes[0]

	fee, err = n.client.EstimateFee(numBlocks)
	if err != nil {
		return 0, err
	}

	return fee, nil
}

//InitMempool makes mempool usable by sending transaction and generating blocks.
func (b *Bitbox) InitMempool() (err error) {
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
