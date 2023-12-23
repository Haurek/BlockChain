package blockchain

import (
	"BlockChain/src/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Block type represents a block in the blockchain
type Block struct {
	Header             *BlockHeader   `json:"header"`             // Header of the block
	TransactionCounter int            `json:"transactionCounter"` // Number of transactions in the block
	Transactions       []*Transaction `json:"transactions"`       // List of transactions in the block
}

// BlockHeader represents the header of a block
type BlockHeader struct {
	Timestamp int64  `json:"timestamp"` // Timestamp when the block was created
	Hash      []byte `json:"hash"`      // Hash of the block
	PrevHash  []byte `json:"prevHash"`  // Hash of the previous block
	Height    uint64 `json:"height"`    // Height of the block in the blockchain
	//MerkleRoot []byte // Placeholder for Merkle Root (commented out)
}

// NewBlock creates a new block with the provided data
func NewBlock(preHash []byte, Txs []*Transaction, height uint64) *Block {
	// Create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = preHash
	header.Hash = []byte{}
	header.Height = height

	// Create the block
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

// NewGenesisBlock creates a genesis block with the given address
func NewGenesisBlock(address []byte) (*Block, error) {
	// Check address validity
	if !CheckAddress(address) {
		return nil, errors.New("wrong address")
	}

	// Create header for the genesis block
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = []byte{}
	header.Hash = []byte{}
	header.Height = 1

	// Create a genesis transaction
	genesisTx := NewCoinbaseTx(address, GenesisValue)

	// Serialize Transactions
	Txs := []*Transaction{genesisTx}

	// Create the genesis block
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

// IsGenesisBlock checks if the block is a genesis block
func (block *Block) IsGenesisBlock() bool {
	return block.Header.Height == 1
}

// Show prints the details of the block to the console
func (block *Block) Show() {
	fmt.Println("Hash: ", hex.EncodeToString(block.Header.Hash))
	fmt.Println("Height: ", block.Header.Height)
	fmt.Println("PreHash: ", hex.EncodeToString(block.Header.PrevHash))
	fmt.Println("Timestamp: ", block.Header.Timestamp)
}
