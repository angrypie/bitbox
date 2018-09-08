package bitbox

//Bitbox represent set API interface to multiple bitcoin nodes,
//that are running in regtest mode.
type Bitbox struct {
	started     bool
	numberNodes int
	nodes       []*bitcoindNode
}

//New creates new Bitbox client
func New() (bitbox Bitbox) {
	return Bitbox{}
}
