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
func (b *Block) Show() {
	fmt.Println("--------------------------------------Block Information--------------------------------------")
	fmt.Printf("Block Information:\n")
	fmt.Printf("  Header:\n")
	fmt.Printf("    Timestamp: %d\n", b.Header.Timestamp)
	fmt.Printf("    Hash: %s\n", hex.EncodeToString(b.Header.Hash))
	fmt.Printf("    PrevHash: %s\n", hex.EncodeToString(b.Header.PrevHash))
	fmt.Printf("    Height: %d\n", b.Header.Height)
	fmt.Println("-----------------------------------Transactions Information----------------------------------")
	fmt.Printf("  TransactionCounter: %d\n", b.TransactionCounter)
	fmt.Printf("  Transactions:\n")
	fmt.Println("---------------------------------------------------------------------------------------------")
	for i, tx := range b.Transactions {
		fmt.Printf("    Transaction %d:\n", i+1)
		fmt.Println("    Inputs:")
		for i, input := range tx.Inputs {
			fmt.Printf("      Inputs %d:\n", i)
			fmt.Printf("        TxID: %s\n", hex.EncodeToString(input.TxID))
			fmt.Printf("        Index: %d\n", input.Index)
			fmt.Printf("        FromAddress: %s\n\n", string(input.FromAddress))
		}
		fmt.Println("    Outputs:")
		for j, output := range tx.Outputs {
			fmt.Printf("      Outputs: %d:\n", j)
			fmt.Printf("        Value: %d\n", output.Value)
			fmt.Printf("        ToAddress: %s\n\n", string(output.ToAddress))
		}
		fmt.Println("---------------------------------------------------------------------------------------------")

	}
}
