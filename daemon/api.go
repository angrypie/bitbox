package daemon

import (
	api "github.com/angrypie/bitbox/proto"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"golang.org/x/net/context"
)

func statusResp(status bool, msg string) (*api.StatusResponse, error) {
	return &api.StatusResponse{
		Status: status,
		Msg:    msg,
	}, nil
}

func (b *Bitbox) Start(ctx context.Context, req *api.StartRequest) (*api.StatusResponse, error) {
	//TODO chekc is it safe
	nodes := int(req.GetNodes())
	if nodes < 1 {
		return statusResp(false, "Number of nodes should be greater than 0")
	}
	b.started = true
	b.numberNodes = nodes

	for i := 0; i < int(nodes); i++ {
		node := startNode(i)
		b.nodes = append(b.nodes, node)
	}

	return statusResp(true, "Success")
}

func (b *Bitbox) Stop(ctx context.Context, req *api.Empty) (*api.StatusResponse, error) {
	//Stop all nodes
	for _, node := range b.nodes {
		node.Stop()
	}
	return statusResp(true, "Success")
}

func (b *Bitbox) Generate(ctx context.Context, req *api.GenerateRequest) (*api.StatusResponse, error) {
	nodeIndex := int(req.GetNode())
	blocks := uint32(req.GetBlocks())
	node := b.nodes[nodeIndex]

	_, err := node.client.Generate(blocks)
	if err != nil {
		return statusResp(false, err.Error())
	}
	return statusResp(true, "Success")
}

func (b *Bitbox) Send(ctx context.Context, req *api.SendRequest) (*api.StatusResponse, error) {
	nodeIndex := int(req.GetNode())
	address := req.GetAddress()

	addr, err := btcutil.DecodeAddress(address, &chaincfg.RegressionNetParams)
	if err != nil {
		return statusResp(false, "Wrong addres: "+err.Error())
	}
	amount, err := btcutil.NewAmount(req.GetAmount())
	if err != nil {
		return statusResp(false, "Wrong amount: "+err.Error())
	}

	node := b.nodes[nodeIndex]
	result := node.client.SendFromAsync("", addr, amount)
	hash, err := result.Receive()
	if err != nil {
		return statusResp(false, err.Error())
	}

	return statusResp(true, hash.String())
}

func (b *Bitbox) Balance(ctx context.Context, req *api.NodeRequest) (*api.BalanceResponse, error) {
	nodeIndex := int(req.GetNode())
	node := b.nodes[nodeIndex]

	amount, err := node.client.GetBalance("")
	if err != nil {
		return nil, err
	}

	return &api.BalanceResponse{Balance: amount.ToBTC()}, nil
}

func (b *Bitbox) Address(ctx context.Context, req *api.NodeRequest) (*api.AddressResponse, error) {
	nodeIndex := int(req.GetNode())
	node := b.nodes[nodeIndex]

	address, err := node.client.GetNewAddress("")
	if err != nil {
		return nil, err
	}

	return &api.AddressResponse{Address: address.String()}, nil
}
