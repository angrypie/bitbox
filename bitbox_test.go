package bitbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitbox(t *testing.T) {
	bb := New()
	assert.Equal(t, bb.started, false)
	assert.Equal(t, bb.numberNodes, 0)
	assert.Len(t, bb.nodes, 0)

}
