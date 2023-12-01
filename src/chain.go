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

// FindEnoughUTXO search unspent output on the chain
func (chain *Chain) FindEnoughUTXO(wallet *Wallet) (int, []*UTXO) {
	// TODO
	return 0, nil
}

// FindTransaction search a transaction by ID
func (chain *Chain) FindTransaction(id []byte) (*Transaction, error) {
	// TODO
	return nil, nil
}
