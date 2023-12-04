package main

import (
	"badger"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
)

// Chain type
type Chain struct {
	Tip        []byte
	BestHeight int
	DataBase   *badger.DB
	Lock       sync.Mutex
}

// BlockIterator block iterator
type BlockIterator struct {
	CurrentHash []byte
	DataBase    *badger.DB
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
	var currentBlock Block
	err = Deserialize(serializeData, &currentBlock)
	HandleError(err)
	// update next point
	iterator.CurrentHash = currentBlock.Header.PrevHash

	return &currentBlock
}

// CreateChain create a new chain
// address use for genesis block
func CreateChain(address []byte) *Chain {
	// create a database
	_, err := os.Stat(DataBaseFile)
	if !os.IsNotExist(err) {
		fmt.Println("Chain database already exists")
		// load database
		return LoadChain()
	}

	// open database
	opts := badger.DefaultOptions(DataBasePath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	HandleError(err)

	// Genesis node create GenesisBlock
	genesisBlock, err := NewGenesisBlock(address)
	HandleError(err)

	// add latest block flag
	err = WriteToDB(db, []byte(BlockTable), []byte(TipHashKey), genesisBlock.Header.Hash)
	HandleError(err)

	chain := &Chain{
		Tip:        genesisBlock.Header.Hash,
		BestHeight: 1,
		DataBase:   db,
	}
	// add genesis block
	chain.AddGenesisBlock(genesisBlock)

	return chain
}

// LoadChain initialize chain from database
func LoadChain() *Chain {
	// check chain database exist
	if _, err := os.Stat(DataBaseFile); os.IsNotExist(err) {
		fmt.Println("No Chain found")
		return nil
	}
	// load local database
	opts := badger.DefaultOptions(DataBasePath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	HandleError(err)

	// read tip hash from database
	latestHash, err := ReadFromDB(db, []byte(BlockTable), []byte(TipHashKey))
	HandleError(err)

	chain := &Chain{
		Tip:      latestHash,
		DataBase: db,
	}

	return chain
}

// AddGenesisBlock add genesis block to chain and update UTXO set
func (chain *Chain) AddGenesisBlock(genesisBlock *Block) {
	// add genesis block to Chain
	data, err := Serialize(genesisBlock)
	HandleError(err)
	err = WriteToDB(chain.DataBase, []byte(BlockTable), genesisBlock.Header.Hash, data)
	HandleError(err)

	// add GenesisBlock output to UTXO set
	genesisOutput := genesisBlock.Transactions[0].Outputs[0]
	var utxos []UTXO
	utxos = append(utxos, UTXO{
		Index:  0,
		Output: genesisOutput,
	})
	data, err = Serialize(utxos)
	HandleError(err)
	err = WriteToDB(chain.DataBase, []byte(ChainStateTable), genesisBlock.Transactions[0].ID, data)
	HandleError(err)
}

// AddBlock add a block to chain
func (chain *Chain) AddBlock(block *Block) bool {
	// get current block previous hash
	preHash := block.Header.PrevHash

	// get previous block from database
	preBlockData, err := ReadFromDB(chain.DataBase, []byte(BlockTable), preHash)
	HandleError(err)
	var preBlock Block
	err = Deserialize(preBlockData, &preBlock)
	HandleError(err)

	// check hash and PoW verify
	if bytes.Equal(preBlock.Header.Hash, block.Header.PrevHash) && Proof(block.Header) {
		// add to database
		serializeData, err := Serialize(block)
		HandleError(err)
		err = WriteToDB(chain.DataBase, []byte(BlockTable), block.Header.Hash, serializeData)
		HandleError(err)

		// update UTXO set
		UpdateUTXOSet(chain.DataBase, block)

		// update tip
		chain.Tip = block.Header.Hash
		chain.BestHeight += 1
		return true
	}

	return false
}

// HaveBlock check a block in chain
func (chain *Chain) HaveBlock(hash []byte) bool {
	// check block is valid
	if chain.FindBlock(hash) != nil {
		return true
	}
	return false
}

// FindBlock return a block by hash
func (chain *Chain) FindBlock(hash []byte) *Block {
	serializeData, err := ReadFromDB(chain.DataBase, []byte(BlockTable), hash)
	HandleError(err)
	var block Block
	err = Deserialize(serializeData, &block)
	HandleError(err)
	return &block
}

// FindTransaction return a transaction by ID
func (chain *Chain) FindTransaction(id []byte) (*Transaction, error) {
	iter := chain.Iterator()
	for {
		block := iter.Next()

		for _, Tx := range block.Transactions {
			if bytes.Equal(Tx.ID, id) {
				return Tx, nil
			}
		}
		// check Genesis Block
		if block.IsGenesisBlock() {
			break
		}
	}
	return nil, errors.New("Transaction not found")
}

// FindUTXO find all UTXO in Chain
func (chain *Chain) FindUTXO() map[string][]UTXO {
	utxosMap := make(map[string][]UTXO)
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
					utxos := utxosMap[id]
					utxo := UTXO{
						Index:  outIndex,
						Output: out,
					}
					utxos = append(utxos, utxo)
					utxosMap[id] = utxos
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
	return utxosMap
}
