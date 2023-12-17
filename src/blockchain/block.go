package blockchain

import (
	"BlockChain/src/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Block type
type Block struct {
	Header             *BlockHeader   `json:"header"`
	TransactionCounter int            `json:"transactionCounter"`
	Transactions       []*Transaction `json:"transactions"`
}

// BlockHeader header of a block
type BlockHeader struct {
	Timestamp int64  `json:"timestamp"`
	Hash      []byte `json:"hash"`
	PrevHash  []byte `json:"prevHash"`
	Height    uint64 `json:"height"`
	//MerkleRoot []byte
}

// NewBlock create a new block
func NewBlock(preHash []byte, Txs []*Transaction, height uint64) *Block {
	// create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = preHash
	header.Hash = []byte{}
	header.Height = height

	block := Block{
		Header:             header,
		Transactions:       Txs,
		TransactionCounter: len(Txs),
	}
	blockBytes, err := utils.Serialize(&block)
	utils.HandleError(err)
	block.Header.Hash = utils.Sha256Hash(blockBytes)

	return &block
}

// NewGenesisBlock create a genesis block
func NewGenesisBlock(address []byte) (*Block, error) {
	// check address
	if !CheckAddress(address) {
		return nil, errors.New("wrong address")
	}
	// create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = []byte{}
	header.Hash = []byte{}
	header.Height = 1
	genesisTx := NewCoinbaseTx(address, GenesisValue)

	// Serialize Transactions
	Txs := []*Transaction{genesisTx}

	genesisBlock := Block{
		Header:             header,
		Transactions:       Txs,
		TransactionCounter: len(Txs),
	}
	genesisBlockBytes, err := utils.Serialize(genesisBlock)
	utils.HandleError(err)
	genesisBlock.Header.Hash = utils.Sha256Hash(genesisBlockBytes)
	return &genesisBlock, nil
}

// IsGenesisBlock check if the block is genesis block
func (block *Block) IsGenesisBlock() bool {
	return block.Header.Height == 1
}

func (block *Block) Show() {
	fmt.Println("Hash: ", hex.EncodeToString(block.Header.Hash))
	fmt.Println("Height: ", block.Header.Height)
	fmt.Println("PreHash: ", hex.EncodeToString(block.Header.PrevHash))
	fmt.Println("Timestamp: ", block.Header.Timestamp)
}
