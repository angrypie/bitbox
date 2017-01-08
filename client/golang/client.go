package bitbox

import (
	"context"
	"errors"
	"github.com/angrypie/bitbox/daemon"
	proto "github.com/angrypie/bitbox/proto"
	"google.golang.org/grpc"
)

type Client struct {
	api proto.BitboxClient
}

// Method only for go package because daemon writen in Golang
func StartDaemon(port int) *grpc.Server {
	return daemon.Start(port)
}

func New(addr string) (*Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	client := proto.NewBitboxClient(conn)
	return &Client{
		api: client,
	}, nil
}

//TODO remove msg return value ??
func (client *Client) Start(nodes int) (status bool, msg string) {
	resp, err := client.api.Start(context.Background(), &proto.StartRequest{Nodes: int32(nodes)})
	if err != nil {
		return false, err.Error()
	}
	return resp.GetStatus(), resp.GetMsg()
}

func (client *Client) Stop() (status bool, msg string) {
	resp, err := client.api.Stop(context.Background(), &proto.Empty{})
	if err != nil {
		return false, err.Error()
	}
	return resp.GetStatus(), resp.GetMsg()
}

func (client *Client) Generate(node, blocks int) (status bool, msg string) {
	req := &proto.GenerateRequest{Node: int32(node), Blocks: int32(blocks)}
	resp, err := client.api.Generate(context.Background(), req)
	if err != nil {
		return false, err.Error()
	}
	return resp.GetStatus(), resp.GetMsg()
}

func (client *Client) Balance(node int) (balance *float64) {
	req := &proto.NodeRequest{Node: int32(node)}
	resp, err := client.api.Balance(context.Background(), req)
	if err != nil {
		return nil
	}
	b := btcRound(resp.GetBalance())
	return &b
}

func (client *Client) Address(node int) (address *string) {
	req := &proto.NodeRequest{Node: int32(node)}
	resp, err := client.api.Address(context.Background(), req)
	if err != nil {
		return nil
	}
	a := resp.GetAddress()
	return &a
}

func (client *Client) Send(node int, address string, amount float64) (tx *string) {
	req := &proto.SendRequest{Node: int32(node), Address: address, Amount: amount}
	resp, err := client.api.Send(context.Background(), req)
	if err != nil {
		return nil
	}
	m := resp.GetMsg()
	return &m
}

func btcRound(b float64) float64 {
	//TODO: refactor
	return float64(int((b * 10000000) / 10000000))
}
