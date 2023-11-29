package BlockChain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

// Block type
type Block struct {
	Header             *BlockHeader
	TransactionCounter int
	Transactions       []*Transaction
}

// BlockHeader header of a block
type BlockHeader struct {
	Timestamp  int64
	Hash       []byte
	PrevHash   []byte
	Nonce      int64
	Bits       int
	MerkelRoot []byte
}

// NewBlock create a new block
func (block *Block) NewBlock(preHash []byte, Txs []*Transaction) *Block {
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = preHash
	header.Bits = 8 // default

	// merkel tree root get from merkel util
	var TxHashs [][]byte
	for _, Tx := range Txs {
		hash := Tx.Hash()
		TxHashs = append(TxHashs, hash)
	}
	merkerTree := &MerkelTree{}
	merkerTree.NewMerkelTree(TxHashs)
	header.MerkelRoot = merkerTree.Root.Hash

	// hash and nonce get from pow util
	header.Nonce, header.Hash = Mining(header)

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
func (block *Block) Deserialize(raw []byte) *Block {
	// TODO
	return nil
}

// Serialize BlockHeader struct
func (header *BlockHeader) Serialize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(header)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return buf.Bytes()
}

// Deserialize []byte data to BlockHeader type
func (header *BlockHeader) Deserialize(raw []byte) *BlockHeader {
	decoder := gob.NewDecoder(bytes.NewReader(raw))
	err := decoder.Decode(&header)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return header
}
