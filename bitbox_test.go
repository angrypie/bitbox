package bitbox

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BitboxTestSuite struct {
	suite.Suite
	backendName string
	b           Bitbox
}

func (suite *BitboxTestSuite) TearDownSuite() {
	suite.NoError(suite.b.Stop())
}

func (suite *BitboxTestSuite) TestBitbox() {
	//Check New, .Start, .Stop
	b := New(suite.backendName)
	suite.b = b
	require := require.New(suite.T())
	state := b.Info()
	require.Equal(false, state.IsStarted)
	require.Equal(0, state.NodesNumber)

	require.NoError(b.Start(2))
	require.NoError(b.InitMempool())

	//Test Info method
	state = b.Info()
	require.True(state.IsStarted)
	require.NotEmpty(state.NodePort)
	require.NotEmpty(state.ZmqAddress)

	//Check regtest methods .Generate .Send
	require.NoError(b.Generate(0, 101))
	time.Sleep(time.Millisecond * 200)

	balance, err := b.Balance(0)
	require.NoError(err)
	log.Println(balance)
	require.True(balance > float64(50), "Balance expected to be greater than 50")

	address, err := b.Address(1)
	require.NoError(err)
	require.NotEmpty(address)

	tx, err := b.Send(0, address, 0.18)
	require.NoError(err)
	require.NotEmpty(tx)

	checkBalance, err := b.Balance(1)
	require.NoError(err)
	require.Equal(float64(0), checkBalance, "Expected balance equal to 0 before TX confirmation")

	require.NoError(b.Generate(0, 3))
	//TODO smart waiting loop
	time.Sleep(time.Millisecond * 5000)

	checkBalance, err = b.Balance(1)
	require.NoError(err)
	require.Equal(0.18, checkBalance, "Expected balance equal to 0.18 after TX confirmation")

	result, err := b.GetRawTransaction(tx)
	require.NoError(err)
	require.Equal(tx, result.Hash().String())

	height, err := b.BlockHeight()
	require.NoError(err)
	require.NotZero(height)
}

func TestBtcd(t *testing.T) {
	suite.Run(t, &BitboxTestSuite{backendName: "btcd"})
}

func TestBitcoind(t *testing.T) {
	suite.Run(t, &BitboxTestSuite{backendName: "bitcoind"})
}
