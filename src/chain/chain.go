// Package chain
package chain

import (
	"BlockChain/src/UTXO"
	"BlockChain/src/Wallet"
	"BlockChain/src/block"
)

type Chain struct {
	blocks []*block.Block
}

// CreateChain create a new chain
func (chain *Chain) CreateChain() *Chain {
	// TODO
	return chain
}

// AddBlock add a block to chain
func (chain *Chain) AddBlock(block *block.Block) bool {
	// TODO
	return true
}

// FindEnoughUnspentOutput search unspent output on the chain
func FindEnoughUnspentOutput(wallet *Wallet.Wallet) (balance int, outputs []*UTXO.TXoutput) {
	// TODO
	return 0, nil
}

// GetPreTransaction search a UTXO input previous transaction on the chain
func GetPreTransaction(input *UTXO.TXinput) *UTXO.Transaction {
	// TODO
	return nil
}
