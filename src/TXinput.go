package BlockChain

import "bytes"

// TXinput is transaction input type
type TXinput struct {
	TxID           []byte
	Value          int
	FromAddress    []byte
	Signature      []byte
	PublicKeyBytes []byte
}

// NewTXinput create new transaction input
func NewTXinput(value int, from, id, pubKey []byte) *TXinput {
	input := &TXinput{
		TxID:           id,
		Value:          value,
		FromAddress:    from,
		Signature:      nil,
		PublicKeyBytes: pubKey,
	}
	return input
}

func (input *TXinput) SetSignature(sig []byte) {
	input.Signature = sig
}

// CanUnlock return an output key hash equal input key hash
func (input *TXinput) CanUnlock(KeyHash []byte) bool {
	return bytes.Equal(KeyHash, Address2PublicKeyHash(input.FromAddress))
}
