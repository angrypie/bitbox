
Bitbox is a golang library that utilize bitcoind regtest mode to create environment for testing your apps.

[![GoDoc](https://godoc.org/github.com/angrypie/bitbox?status.svg)](https://godoc.org/github.com/angrypie/bitbox)


## Quick start

Use example below to bootstrap your tests.


```go
import "github.com/angrypie/bitbox"

func TestSomething(t *testing.T) {
	client := bitbox.New()
	//Start 2 bitcoind nodes connected together
	err := client.Start(2) // you should have bitcoind installed
	if err != nil {
		t.Error(err)
	}
	blocks, _ := client.BlockHeight()        // get current network block height
	client.Generate(0, 105)                  // generate first blocks, reward will go to 0 account
	address, _ := client.Address(1)          // get new address of second node
	tx, _ := client.Send(0, address, 0.0001) // get new address of second node
	client.Generate(0, 1)                    // generate another block to get confirmation

	log.Println(block, address, tx)
	// In real world don't forget to check errors
}


