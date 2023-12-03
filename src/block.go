package BlockChain

import (
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
func NewBlock(preHash []byte, Txs []*Transaction) *Block {
	// create header
	header := &BlockHeader{}
	header.Timestamp = time.Now().Unix()
	header.PrevHash = preHash
	header.Bits = Difficulty

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
func NewGenesisBlock(address []byte) *Block {
	coinBaseTx := NewCoinbaseTx(address)
	return NewBlock([]byte{}, []*Transaction{coinBaseTx})
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.Header.PrevHash) == 0
}
