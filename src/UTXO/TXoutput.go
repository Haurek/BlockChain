package UTXO

// TXoutput is UTXO output type
type TXoutput struct {
	Index   int
	Value   int
	Address []byte
}

// NewTXoutput create new UTXO output
func (output *TXoutput) NewTXoutput(index int, value int, address []byte) *TXoutput {
	output.Index = index
	output.Value = value
	output.Address = address
	return output
}
