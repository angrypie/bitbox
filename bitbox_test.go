package bitbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitbox(t *testing.T) {
	//Check New, .Start, .Stop
	b := New()
	assert.Equal(t, false, b.started)
	assert.Equal(t, 0, b.numberNodes)
	assert.Len(t, b.nodes, 0)

	require.Nil(t, b.Start(2))
	defer func() {
		assert.Nil(t, b.Stop())
	}()
	require.Nil(t, b.InitMempool())

	//Check regtest methods .Generate .Send
	require.Nil(t, b.Generate(0, 101))
	time.Sleep(time.Millisecond * 200)

	balance, err := b.Balance(0)
	require.Nil(t, err)
	assert.True(t, balance > float64(50))

	address, err := b.Address(1)
	require.Nil(t, err)
	assert.NotEmpty(t, address)

	tx, err := b.Send(0, address, 0.18)
	require.Nil(t, err)
	assert.NotEmpty(t, tx)

	checkBalance, err := b.Balance(1)
	require.Nil(t, err)
	assert.Equal(t, float64(0), checkBalance, "Expected balance equal to 0 before TX confirmation")

	require.Nil(t, b.Generate(0, 1))
	time.Sleep(time.Millisecond * 200)

	checkBalance, err = b.Balance(1)
	require.Nil(t, err)
	assert.Equal(t, 0.18, checkBalance, "Expected balance equal to 0.18 after TX confirmation")

	result, err := b.GetRawTransaction(tx)
	require.Nil(t, err)
	assert.Equal(t, tx, result.Hash().String())

}
