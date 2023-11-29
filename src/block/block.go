// Package block
package block

import (
	"BlockChain/src/UTXO"
	"time"
)

// Block type
type Block struct {
	Header       *BlockHeader
	Transactions []*UTXO.Transaction
}

// BlockHeader header of a block
type BlockHeader struct {
	Timestamp  int64
	Hash       [32]byte
	PrevHash   [32]byte
	Nonce      uint
	Height     uint
	MerkelRoot string // temporary
}

// NewBlock create a new block
func (block *Block) NewBlock(prehash [32]byte, Txs []*UTXO.Transaction, height uint) *Block {
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()

	// hash and nonce get from pow util
	header.Nonce = 1234
	header.Hash = [32]byte(make([]byte, 32))
	header.PrevHash = prehash
	header.Height = height

	// merkel tree root get from merkel util
	header.MerkelRoot = "root"

	block.Header = header
	block.Transactions = Txs
	return block
}

// Serialize Block struct
func (block *Block) Serialize() []byte {
	// TODO
	return nil
}

// Deserialize []byte data to Block type
func (Tx *Block) Deserialize(raw []byte) *Block {
	// TODO
	return nil
}
