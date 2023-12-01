package BlockChain

// TXoutput is UTXO output type
type TXoutput struct {
	Value         int
	ToAddress     []byte
	PublicKeyHash []byte
}

// NewTXoutput create new transaction output
func NewTXoutput(value int, address []byte) *TXoutput {
	if value < 0 {
		panic("wrong value")
	}
	output := &TXoutput{}
	output.Value = value
	output.ToAddress = address
	output.Lock(address)

	return output
}

// Lock sender lock output use address
func (output *TXoutput) Lock(address []byte) []byte {
	output.PublicKeyHash = Address2PublicKeyHash(address)
	return output.PublicKeyHash
}
