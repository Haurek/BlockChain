package main

import (
	"encoding/hex"
	"errors"
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
	MerkleRoot []byte
}

// NewBlock create a new block
func NewBlock(preHash []byte, Txs []*Transaction, minerAddress []byte) *Block {
	// create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = preHash
	header.Bits = Difficulty

	// add coinbase Tx for miner
	coinBaseTx := NewCoinbaseTx(minerAddress, MinerReward)
	Txs = append([]*Transaction{coinBaseTx}, Txs...)

	// Serialize Transactions
	var TxsBytes [][]byte
	for _, Tx := range Txs {
		TxBytes, err := Serialize(Tx)
		HandleError(err)
		TxsBytes = append(TxsBytes, TxBytes)
	}

	// generate merkle tree
	merkleTree := NewMerkleTree(TxsBytes)
	header.MerkleRoot = merkleTree.Root.Hash

	// hash and nonce get from pow util
	header.Nonce, header.Hash = Mining(header)

	block := &Block{
		Header:             header,
		Transactions:       Txs,
		TransactionCounter: len(Txs),
	}
	return block
}

// NewGenesisBlock create a genesis block
func NewGenesisBlock(address []byte) (*Block, error) {
	// check address
	if !IsValidAddress(address) {
		return nil, errors.New("wrong address")
	}
	// create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = []byte{}
	header.Bits = Difficulty
	genesisTx := NewCoinbaseTx(address, GenesisValue)

	// Serialize Transactions
	Txs := []*Transaction{genesisTx}
	var TxsBytes [][]byte
	TxBytes, err := Serialize(genesisTx)
	HandleError(err)
	TxsBytes = append(TxsBytes, TxBytes)

	// generate merkle tree
	merkleTree := NewMerkleTree(TxsBytes)
	header.MerkleRoot = merkleTree.Root.Hash

	// hash and nonce get from pow util
	header.Nonce, header.Hash = Mining(header)

	genesisBlock := &Block{
		Header:             header,
		Transactions:       Txs,
		TransactionCounter: len(Txs),
	}
	return genesisBlock, nil
}

// IsGenesisBlock check if the block is genesis block
func (block *Block) IsGenesisBlock() bool {
	return len(block.Header.PrevHash) == 0
}

// Show block message
func (block *Block) Show() {
	fmt.Println("Block head:")
	fmt.Println("	Hash: ", hex.EncodeToString(block.Header.Hash))
	fmt.Println("	previous Hash: ", hex.EncodeToString(block.Header.Hash))
	fmt.Println("	Time: ", time.Unix(block.Header.Timestamp, 0))
	fmt.Println("	Nonce: ", block.Header.Nonce)
	fmt.Println("	Bits: ", block.Header.Bits)
	fmt.Println("	MerkleTree Root: ", hex.EncodeToString(block.Header.MerkleRoot))
	fmt.Println("Transactions:")
	// TODO
}
