package blockchain

// TXinput is transaction input type
// TxID is previous Tx ID
// Index is the index of output corresponding to the input
type TXinput struct {
	TxID           []byte `json:"txID"`
	Index          int    `json:"index"`
	FromAddress    []byte `json:"fromAddress"`
	Signature      []byte `json:"signature"`
	PublicKeyBytes []byte `json:"publicKeyBytes"`
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
