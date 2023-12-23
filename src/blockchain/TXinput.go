package blockchain

// TXinput represents the structure of a transaction input
// TxID refers to the ID of the previous transaction
// Index is the index of the output corresponding to the input
type TXinput struct {
	TxID           []byte `json:"txID"`           // TxID is the ID of the previous transaction
	Index          int    `json:"index"`          // Index represents the index of the output corresponding to the input
	FromAddress    []byte `json:"fromAddress"`    // FromAddress denotes the address from where the input comes
	Signature      []byte `json:"signature"`      // Signature is the digital signature of the input
	PublicKeyBytes []byte `json:"publicKeyBytes"` // PublicKeyBytes represents the public key bytes of the sender
}

// NewTXinput creates a new transaction input with given index, fromAddress, transaction ID, and public key
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

// SetSignature sets the signature for a transaction input
func (input *TXinput) SetSignature(sig []byte) {
	input.Signature = sig
}
