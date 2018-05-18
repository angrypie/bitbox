package bitbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitbox(t *testing.T) {
	b := New()
	assert.Equal(t, b.started, false)
	assert.Equal(t, b.numberNodes, 0)
	assert.Len(t, b.nodes, 0)

	require.Nil(t, b.Start(2))
	defer func() {
		assert.Nil(t, b.Stop())
	}()

}
