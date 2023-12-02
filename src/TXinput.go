package BlockChain

// TXinput is transaction input type
// TxID is previous Tx ID
// Index is the index of output corresponding to the input
type TXinput struct {
	TxID           []byte
	Index          int
	FromAddress    []byte
	Signature      []byte
	PublicKeyBytes []byte
}

// NewTXinput create new transaction input
func NewTXinput(index int, from, id, pubKey []byte) *TXinput {
	input := &TXinput{
		TxID:           id,
		Index:          index,
		FromAddress:    from,
		Signature:      nil,
		PublicKeyBytes: pubKey,
	}
	return input
}

func (input *TXinput) SetSignature(sig []byte) {
	input.Signature = sig
}
