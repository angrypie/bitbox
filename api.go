package bitbox

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func (b *Bitbox) Start(nodes int) (err error) {
	if nodes < 1 {
		return errors.New("Number of nodes should be greater than 0")
	}

	b.started = true
	b.numberNodes = nodes

	for i := 0; i < int(nodes); i++ {
		node := startNode(i)
		b.nodes = append(b.nodes, node)
	}

	return nil
}

func (b *Bitbox) Stop() (err error) {
	//Stop all nodes
	for _, node := range b.nodes {
		node.Stop()
	}
	return nil
}

func (b *Bitbox) Generate(nodeIndex int, blocks uint32) (err error) {
	node := b.nodes[nodeIndex]

	_, err = node.client.Generate(blocks)
	if err != nil {
		return err
	}

	return nil
}

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
	result := n.client.SendFromAsync("", addr, btcAmount)
	hash, err := result.Receive()
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func (b *Bitbox) Balance(node int) (balance float64, err error) {
	n := b.nodes[node]

	amount, err := n.client.GetBalance("")
	if err != nil {
		return 0, err
	}

	return amount.ToBTC(), nil
}

func (b *Bitbox) Address(node int) (address string, err error) {
	n := b.nodes[node]

	addr, err := n.client.GetNewAddress("")
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}
