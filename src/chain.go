package BlockChain

import (
	badger "badger-4.2.0"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
)

// Chain type
type Chain struct {
	Tip      []byte
	DataBase *badger.DB
}

// BlockIterator block iterator
type BlockIterator struct {
	CurrentHash []byte
	DataBase    *badger.DB
}

// CreateChain create a new chain
// address use for genesis block
func CreateChain(address []byte) *Chain {
	// create a database
	if _, err := os.Stat(DataBaseFile); os.IsNotExist(err) {
		fmt.Println("Chain database already exists")
		// load database
		return InitChain()
	}

	opts := badger.DefaultOptions(DataBasePath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	HandleError(err)

	// Genesis node create GenesisBlock
	coinBaseTx := NewCoinbaseTx(address)
	genesisBlock := NewGenesisBlock(coinBaseTx)
	// add genesis block
	data, err := Serialize(genesisBlock)
	HandleError(err)
	err = WriteToDB(db, []byte(BlockTable), genesisBlock.Header.Hash, data)
	HandleError(err)
	// add latest block flag
	err = WriteToDB(db, []byte(BlockTable), []byte(TipHashKey), genesisBlock.Header.Hash)
	HandleError(err)
	chain := &Chain{
		Tip:      genesisBlock.Header.Hash,
		DataBase: db,
	}
	return chain
}

// InitChain initialize chain from database
func InitChain() *Chain {
	if _, err := os.Stat(DataBaseFile); os.IsNotExist(err) {
		fmt.Println("No Chain found")
		return nil
	}
	opts := badger.DefaultOptions(DataBasePath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	HandleError(err)

	latestHash, err := ReadFromDB(db, []byte(BlockTable), []byte(TipHashKey))
	HandleError(err)

	chain := &Chain{
		Tip:      latestHash,
		DataBase: db,
	}
	return chain
}

// Iterator create a block iterator
func (chain *Chain) Iterator() *BlockIterator {
	iterator := BlockIterator{
		CurrentHash: chain.Tip,
		DataBase:    chain.DataBase,
	}
	return &iterator
}

// Next get previous block by preHash
func (iterator *BlockIterator) Next() *Block {
	// get current block serialize data
	serializeData, err := ReadFromDB(iterator.DataBase, []byte(BlockTable), iterator.CurrentHash)
	HandleError(err)
	// deserialize data
	var currentBlock *Block
	err = Deserialize(serializeData, currentBlock)
	HandleError(err)
	// update next point
	iterator.CurrentHash = currentBlock.Header.PrevHash

	return currentBlock
}

// AddBlock add a block to chain
func (chain *Chain) AddBlock(block *Block) bool {
	// TODO
	return true
}

// HaveBlock check a block in chain
func (chain *Chain) HaveBlock(hash []byte) bool {
	// TODO
	return false
}

// FindBlock return a block by hash
func (chain *Chain) FindBlock(hash []byte) *Block {
	serializeData, err := ReadFromDB(chain.DataBase, []byte(BlockTable), hash)
	HandleError(err)
	var block *Block
	err = Deserialize(serializeData, block)
	HandleError(err)
	return block
}

// FindTransaction return a transaction by ID
func (chain *Chain) FindTransaction(id []byte) (*Transaction, error) {
	iter := chain.Iterator()
	for {
		block := iter.Next()

		// check Genesis Block
		if block.IsGenesisBlock() {
			break
		}
		for _, Tx := range block.Transactions {
			if bytes.Equal(Tx.ID, id) {
				return Tx, nil
			}
		}
	}
	return nil, errors.New("Transaction not found")
}

// FindUTXO find all UTXO in Chain
func (chain *Chain) FindUTXO() map[string]UTXO {
	utxos := make(map[string]UTXO)
	spentTxOutputs := make(map[string][]int)
	iter := chain.Iterator()

	// traverse block
	for {
		block := iter.Next()
		// traverse Transaction
		for _, Tx := range block.Transactions {
			id := hex.EncodeToString(Tx.ID)
			// traverse Outputs
			for outIndex, out := range Tx.Outputs {
				spent := false
				if spentTxOutputs[id] != nil {
					// There is output that has been spent in the Tx
					for _, spentIndex := range spentTxOutputs[id] {
						if spentIndex == outIndex {
							spent = true
							break
						}
					}
				}
				// out is unspent
				if spent == false {
					utxos[id].outputs = append(utxos[id].outputs, out)
				}
			}

			if Tx.IsCoinBase() == false {
				// traverse Outputs
				for _, in := range Tx.Inputs {
					preTxId := hex.EncodeToString(in.TxID)
					spentTxOutputs[preTxId] = append(spentTxOutputs[preTxId], in.Index)
				}
			}
		}

		// end of chain
		if block.IsGenesisBlock() {
			break
		}
	}
	return utxos
}
