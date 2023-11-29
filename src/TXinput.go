package BlockChain

import (
	"bytes"
)

// TXinput is UTXO input type
type TXinput struct {
	//ID 			[]byte // maybe use for search previous Tx
	Index     int
	Value     int
	Address   []byte
	Signature []byte
	PublicKey []byte
}

// NewTXinput create new UTXO input
func (input *TXinput) NewTXinput(index int, value int, wallet *Wallet) *TXinput {
	input.Index = index
	input.Value = value
	input.Address = wallet.Address
	input.PublicKey = wallet.PublicKey
	return input
}

// CheckAddress verify address and public key
func (input *TXinput) CheckAddress(addr []byte) bool {
	pub2addr := GenerateAddress(input.PublicKey)

	return bytes.Compare(addr, pub2addr) == 0
}
