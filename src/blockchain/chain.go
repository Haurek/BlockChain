package blockchain

import (
	"BlockChain/src/utils"
	"badger"
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"sync"
)

// Chain represents the blockchain
type Chain struct {
	Tip        []byte      // Hash of the latest block
	BestHeight uint64      // Height of the best block in the chain
	DataBase   *badger.DB  // Database to store blockchain data
	log        *log.Logger // Logger for the blockchain
	Lock       sync.Mutex  // Mutex for synchronizing access to the blockchain
}

// BlockIterator iterates over blocks in the blockchain
type BlockIterator struct {
	CurrentHash []byte      // Hash of the current block being iterated
	DataBase    *badger.DB  // Database to fetch block data
	Lock        *sync.Mutex // Mutex for synchronizing access to the iterator
}

// Iterator creates a block iterator for the chain
func (chain *Chain) Iterator() *BlockIterator {
	iterator := BlockIterator{
		CurrentHash: chain.Tip,
		DataBase:    chain.DataBase,
		Lock:        &chain.Lock,
	}
	return &iterator
}

// Next gets the previous block by its previous hash
func (iterator *BlockIterator) Next() *Block {
	// Fetch the current block's serialized data
	iterator.Lock.Lock()
	serializeData, err := ReadFromDB(iterator.DataBase, []byte(BlockTable), iterator.CurrentHash)
	iterator.Lock.Unlock()
	utils.HandleError(err)
	// Deserialize the data into a Block
	var currentBlock Block
	err = utils.Deserialize(serializeData, &currentBlock)
	utils.HandleError(err)
	// Update the iterator to point to the previous block
	iterator.CurrentHash = currentBlock.Header.PrevHash

	return &currentBlock
}

// FindBlocksInRange finds all blocks within a given height range in the blockchain
func (chain *Chain) FindBlocksInRange(min, max uint64) []*Block {
	var blocksInRange []*Block

	iter := chain.Iterator()
	for {
		block := iter.Next()

		if block.Header.Height >= min && block.Header.Height <= max {
			// Add the block to the result list if its height is within the specified range
			blocksInRange = append(blocksInRange, block)
		}

		// Stop iterating if the current block is the genesis block
		if block.IsGenesisBlock() {
			break
		}
	}

	return blocksInRange
}

// CreateChain creates a new blockchain with a genesis block
func CreateChain(address []byte, path, logPath string) (*Chain, error) {
	// Initialize logger
	l := utils.NewLogger("[chain] ", logPath)

	// Create a database
	db, err := OpenDatabase(path)
	if err != nil {
		l.Panic("fail to create database")
		return nil, err
	}

	// Create the genesis block
	genesisBlock, err := NewGenesisBlock(address)
	if err != nil {
		l.Panic("fail to create genesis Block")
		return nil, err
	}

	// Add the latest block flag to the database
	err = WriteToDB(db, []byte(BlockTable), []byte(TipHashKey), genesisBlock.Header.Hash)
	if err != nil {
		l.Panic("fail to write block data into database")
		return nil, err
	}

	// Initialize the chain with the genesis block
	chain := &Chain{
		Tip:        genesisBlock.Header.Hash,
		BestHeight: 1,
		DataBase:   db,
		log:        l,
	}
	// Add the genesis block to the chain
	chain.AddGenesisBlock(genesisBlock)

	return chain, nil
}

// LoadChain initializes the blockchain from the database
func LoadChain(path, logPath string) (*Chain, error) {
	// Initialize logger
	l := utils.NewLogger("[chain] ", logPath)

	// Create a database
	db, err := OpenDatabase(path)
	if err != nil {
		l.Panic("fail to create database")
		return nil, err
	}

	// Read the tip hash from the database
	latestHash, err := ReadFromDB(db, []byte(BlockTable), []byte(TipHashKey))

	// Initialize an empty chain if no tip hash is found
	var chain *Chain
	if err != nil {
		chain = &Chain{
			Tip:        nil,
			BestHeight: 0,
			DataBase:   db,
			log:        l,
		}
		return chain, nil
	}

	// Retrieve the best height
	serializeData, err := ReadFromDB(db, []byte(BlockTable), latestHash)
	var heightBlock Block
	err = utils.Deserialize(serializeData, &heightBlock)
	if err != nil {
		l.Panic("fail to read Tip Block")
		return nil, err
	}
	chain = &Chain{
		Tip:        latestHash,
		BestHeight: heightBlock.Header.Height,
		DataBase:   db,
		log:        l,
	}

	return chain, nil
}

// ... (previous functions continued)

// AddGenesisBlock adds the genesis block to the chain and updates the UTXO set
func (chain *Chain) AddGenesisBlock(genesisBlock *Block) {
	// Add the genesis block to the chain in the database
	data, err := utils.Serialize(genesisBlock)
	utils.HandleError(err)
	chain.Lock.Lock()
	err = WriteToDB(chain.DataBase, []byte(BlockTable), genesisBlock.Header.Hash, data)
	chain.Lock.Unlock()
	utils.HandleError(err)

	// Add the GenesisBlock output to the UTXO set
	genesisOutput := genesisBlock.Transactions[0].Outputs[0]
	var utxos []UTXO
	utxos = append(utxos, UTXO{
		Index:  0,
		Output: genesisOutput,
	})
	data, err = utils.Serialize(utxos)
	utils.HandleError(err)
	chain.Lock.Lock()
	err = WriteToDB(chain.DataBase, []byte(ChainStateTable), genesisBlock.Transactions[0].ID, data)
	chain.Lock.Unlock()
	utils.HandleError(err)
}

// AddBlock adds a block to the chain
func (chain *Chain) AddBlock(block *Block) bool {
	// Retrieve the previous block's hash from the current block
	preHash := block.Header.PrevHash

	// If the new block is a genesis block
	if block.IsGenesisBlock() {
		// Add the genesis block to the chain
		chain.AddGenesisBlock(block)
		// Update chain metadata for the genesis block
		chain.Lock.Lock()
		chain.Tip = block.Header.Hash
		chain.BestHeight = block.Header.Height
		err := WriteToDB(chain.DataBase, []byte(BlockTable), []byte(TipHashKey), block.Header.Hash)
		if err != nil {
			defer chain.Lock.Unlock()
			chain.log.Println("Update tip fail")
			return false
		}
		chain.Lock.Unlock()

		return true
	}
	// get previous block from database
	chain.Lock.Lock()
	preBlockData, err := ReadFromDB(chain.DataBase, []byte(BlockTable), preHash)
	if err != nil {
		defer chain.Lock.Unlock()
		chain.log.Println("Add block to database fail")
		return false
	}
	chain.Lock.Unlock()

	var preBlock Block
	err = utils.Deserialize(preBlockData, &preBlock)
	if err != nil {
		chain.log.Println("Deserialize pre-block fail")
		return false
	}

	if bytes.Equal(preBlock.Header.Hash, block.Header.PrevHash) {
		// add to database
		serializeData, err := utils.Serialize(block)
		if err != nil {
			chain.log.Println("Serialize block fail")
			return false
		}
		chain.Lock.Lock()
		err = WriteToDB(chain.DataBase, []byte(BlockTable), block.Header.Hash, serializeData)
		if err != nil {
			defer chain.Lock.Unlock()
			chain.log.Println("Add block to database fail")
			return false
		}
		// update tip
		err = WriteToDB(chain.DataBase, []byte(BlockTable), []byte(TipHashKey), block.Header.Hash)
		if err != nil {
			defer chain.Lock.Unlock()
			chain.log.Println("Update tip fail")
			return false
		}
		chain.Tip = block.Header.Hash
		chain.BestHeight += 1
		chain.Lock.Unlock()

		// update UTXO set
		ok := UpdateUTXOSet(chain.DataBase, block)
		if !ok {
			ReindexUTXOSet(chain.DataBase, chain.FindUTXO())
		}
		return true
	}

	return false
}

// HaveBlock checks if a block with a given hash exists in the chain
func (chain *Chain) HaveBlock(hash []byte) bool {
	// Check if the block exists in the chain
	if chain.FindBlock(hash) != nil {
		return true
	}
	return false
}

// FindBlock returns a block by its hash
func (chain *Chain) FindBlock(hash []byte) *Block {
	// Find and return the block from the chain using its hash
	chain.Lock.Lock()
	serializeData, err := ReadFromDB(chain.DataBase, []byte(BlockTable), hash)
	chain.Lock.Unlock()
	if err != nil {
		return nil
	}
	var block Block
	err = utils.Deserialize(serializeData, &block)
	if err != nil {
		return nil
	}
	return &block
}

// FindTransaction returns a transaction by its ID
func (chain *Chain) FindTransaction(id []byte) (*Transaction, error) {
	// Search for a transaction by ID in the blockchain
	iter := chain.Iterator()
	for {
		block := iter.Next()

		for _, Tx := range block.Transactions {
			if bytes.Equal(Tx.ID, id) {
				return Tx, nil
			}
		}

		if block.IsGenesisBlock() {
			break
		}
	}
	return nil, errors.New("Transaction not found")
}

// FindUTXO finds all unspent transaction outputs (UTXOs) in the chain
func (chain *Chain) FindUTXO() map[string][]UTXO {
	// Find and return all unspent transaction outputs (UTXOs) in the chain
	utxosMap := make(map[string][]UTXO)
	spentTxOutputs := make(map[string][]int)
	iter := chain.Iterator()

	// Traverse the blocks in the chain
	for {
		block := iter.Next()

		for _, Tx := range block.Transactions {
			id := hex.EncodeToString(Tx.ID)

			for outIndex, out := range Tx.Outputs {
				spent := false
				if spentTxOutputs[id] != nil {
					for _, spentIndex := range spentTxOutputs[id] {
						if spentIndex == outIndex {
							spent = true
							break
						}
					}
				}
				if !spent {
					utxos := utxosMap[id]
					utxo := UTXO{
						Index:  outIndex,
						Output: out,
					}
					utxos = append(utxos, utxo)
					utxosMap[id] = utxos
				}
			}

			if !Tx.IsCoinBase() {
				for _, in := range Tx.Inputs {
					preTxId := hex.EncodeToString(in.TxID)
					spentTxOutputs[preTxId] = append(spentTxOutputs[preTxId], in.Index)
				}
			}
		}

		if block.IsGenesisBlock() {
			break
		}
	}
	return utxosMap
}

// GetTip returns the tip of the chain
func (chain *Chain) GetTip() string {
	chain.Lock.Lock()
	defer chain.Lock.Unlock()
	return hex.EncodeToString(chain.Tip)
}

// GetHeight returns the height of the chain
func (chain *Chain) GetHeight() uint64 {
	chain.Lock.Lock()
	defer chain.Lock.Unlock()
	return chain.BestHeight
}
