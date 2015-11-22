package toxdynboot

/*
ToxNode is a single possible node candidate. The struct contains all fields that
we could parse from the wiki in the correct format to use with the Tox wrapper
directly.
*/
type ToxNode struct {
	IPv4       string
	IPv6       string
	Port       uint16
	PublicKey  []byte
	Maintainer string
	Location   string
}

/*
String is a basic string representation of the information contained in the
ToxNode.
*/
func (t *ToxNode) String() string {
	return "ToxNode{IPv4:" + t.IPv4 + ",IPv6:" + t.IPv6 + ",Port:" +
		string(t.Port) + ",PublicKey:" + string(t.PublicKey) + ",Maintainer:" +
		t.Maintainer + ",Location:" + t.Location + "}"
}
