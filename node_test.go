package bitbox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	masterIndex, slaveIndex := 0, 1

	master := startNode(masterIndex)
	require.NotNil(t, master)
	defer master.Stop()
	defer master.Clean()

	slave := startNode(slaveIndex)
	require.NotNil(t, slave)
	defer slave.Clean()

}
