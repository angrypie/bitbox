package bitbox

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

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

//waitUntilNodeStart TODO set time limit
func waitUntilNodeStart(node Node) (err error) {
	for i := 0; ; i++ {
		err = node.Client().Ping()
		if err != nil {
			if i == 40 {
				fmt.Println("trying to get info", err)
				i = 0
			}
			time.Sleep(time.Millisecond * 100)
			continue
		}
		break
	}

	return
}

type createNodeFunc = func(index int, masterNodePort string) (Node, error)

func newNodeSet(number int, createNode createNodeFunc) (nodes []Node, err error) {
	var masterNodePort string
	node, err := createNode(0, "")
	if err != nil {
		return
	}

	nodes = append(nodes, node)
	masterNodePort = node.Info().NodePort

	var wg sync.WaitGroup
	wg.Add(number - 1)
	for i := 1; i < int(number); i++ {
		i := i
		go func() {
			node, err = createNode(i, masterNodePort)
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

func initBitboxMempool(b Bitbox) (err error) {
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
