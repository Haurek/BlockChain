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
func NewGenesisBlock(coinBase *Transaction) *Block {
	return NewBlock([]byte{}, []*Transaction{coinBase})
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.Header.PrevHash) == 0
}

//// Serialize Block struct
//func (block *Block) Serialize() []byte {
//	// TODO
//	return nil
//}
//
//// DeserializeBlock []byte data to Block type
//func DeserializeBlock(raw []byte) *Block {
//	// TODO
//	return nil
//}
//
//// Serialize BlockHeader struct
//func (header *BlockHeader) Serialize() []byte {
//	var buf bytes.Buffer
//	encoder := gob.NewEncoder(&buf)
//	err := encoder.Encode(header)
//	HandleError(err)
//
//	return buf.Bytes()
//}
//
//// DeserializeBlockHeader []byte data to BlockHeader type
//func DeserializeBlockHeader(raw []byte) *BlockHeader {
//	var header BlockHeader
//	decoder := gob.NewDecoder(bytes.NewReader(raw))
//	err := decoder.Decode(header)
//	HandleError(err)
//	return &header
//}
