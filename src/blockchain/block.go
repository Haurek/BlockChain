package blockchain

import (
	"BlockChain/src/utils"
	"errors"
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
	Timestamp int64
	Hash      []byte
	PrevHash  []byte
	Height    uint64
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
	// Serialize Transactions
	//var TxsBytes [][]byte
	//for _, Tx := range Txs {
	//	TxBytes, err := utils.Serialize(Tx)
	//	utils.HandleError(err)
	//	TxsBytes = append(TxsBytes, TxBytes)
	//}

	// generate merkle tree
	//merkleTree := NewMerkleTree(TxsBytes)
	//header.MerkleRoot = merkleTree.Root.Hash
	//headerData := bytes.Join(
	//	[][]byte{
	//		header.PrevHash,
	//		header.MerkleRoot,
	//		Int2Bytes(header.Timestamp),
	//	},
	//	[]byte{},
	//)

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
	if !IsValidAddress(address) {
		return nil, errors.New("wrong address")
	}
	// create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = []byte{}
	header.Hash = []byte{}
	header.Height = 0
	genesisTx := NewCoinbaseTx(address, GenesisValue)

	//// Serialize Transactions
	Txs := []*Transaction{genesisTx}
	//var TxsBytes [][]byte
	//TxBytes, err := Serialize(genesisTx)
	//HandleError(err)
	//TxsBytes = append(TxsBytes, TxBytes)
	//
	//// generate merkle tree
	//merkleTree := NewMerkleTree(TxsBytes)
	//header.MerkleRoot = merkleTree.Root.Hash

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
	return block.Header.Height == 0
}
