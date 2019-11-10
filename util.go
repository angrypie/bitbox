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

func waitUntilNodeStart(node Node) (err error) {
	err = node.Client().Ping()
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		err = node.Start()
		if err != nil {
			log.Println("ERR starting daemon", err)
			return
		}

		for i := 0; ; i++ {
			time.Sleep(time.Millisecond * 100)
			err = node.Client().Ping()
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
