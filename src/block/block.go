// Package block
package block

import "time"

// Block type
type Block struct {
	Header       *BlockHeader
	Transactions string // temporary
}

// BlockHeader header of a block
type BlockHeader struct {
	Timestamp  int64
	Hash       string
	PrevHash   string
	Nonce      uint
	Height     uint
	MerkelRoot string // temporary
}

// NewBlock create a new block
func (block *Block) NewBlock(prehash string) *Block {
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	// hash and nonce get from pow util
	header.Nonce = 1234
	header.Hash = "5678"
	header.PrevHash = prehash
	header.Height = 0 // temporary
	// merkel tree root get from merkel util
	header.MerkelRoot = "root"
	block.Header = header
	block.Transactions = "Transactions" // temporary
	return block
}
