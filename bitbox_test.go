package bitbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBitbox(t *testing.T) {
	//Check New, .Start, .Stop
	b := New()
	require := require.New(t)
	state := b.Info()
	require.Equal(false, state.IsStarted)
	require.Equal(0, state.NodesNumber)

	require.NoError(b.Start(2))
	defer func() {
		require.NoError(b.Stop())
	}()
	require.NoError(b.InitMempool())

	//Test Info method
	state = b.Info()
	require.True(state.IsStarted)
	require.NotEmpty(state.RPCPort)
	require.NotEmpty(state.ZmqAddress)

	//Check regtest methods .Generate .Send
	require.NoError(b.Generate(0, 101))
	time.Sleep(time.Millisecond * 200)

	balance, err := b.Balance(0)
	require.NoError(err)
	require.True(balance > float64(50))

	address, err := b.Address(1)
	require.NoError(err)
	require.NotEmpty(address)

	tx, err := b.Send(0, address, 0.18)
	require.NoError(err)
	require.NotEmpty(tx)

	checkBalance, err := b.Balance(1)
	require.NoError(err)
	require.Equal(float64(0), checkBalance, "Expected balance equal to 0 before TX confirmation")

	require.NoError(b.Generate(0, 1))
	time.Sleep(time.Millisecond * 200)

	checkBalance, err = b.Balance(1)
	require.NoError(err)
	require.Equal(0.18, checkBalance, "Expected balance equal to 0.18 after TX confirmation")

	result, err := b.GetRawTransaction(tx)
	require.NoError(err)
	require.Equal(tx, result.Hash().String())
}
