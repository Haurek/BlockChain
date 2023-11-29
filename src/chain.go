package BlockChain

type Chain struct {
	blocks []*Block
}

// CreateChain create a new chain
func (chain *Chain) CreateChain() *Chain {
	// TODO
	return chain
}

// AddBlock add a block to chain
func (chain *Chain) AddBlock(block *Block) bool {
	// TODO
	return true
}

// FindEnoughUnspentOutput search unspent output on the chain
func FindEnoughUnspentOutput(wallet *Wallet) (balance int, outputs []*TXoutput) {
	// TODO
	return 0, nil
}

// GetPreTransaction search a UTXO input previous transaction on the chain
func GetPreTransaction(input *TXinput) *Transaction {
	// TODO
	return nil
}
