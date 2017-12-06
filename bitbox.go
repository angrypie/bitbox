package bitbox

type Bitbox struct {
	started     bool
	numberNodes int
	nodes       []*bitcoindNode
}

func New() (bitbox Bitbox) {
	return Bitbox{}
}
