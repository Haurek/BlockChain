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

// Chain type
type Chain struct {
	Tip        []byte
	BestHeight uint64
	DataBase   *badger.DB
	log        *log.Logger
	Lock       sync.Mutex
}

// BlockIterator block iterator
type BlockIterator struct {
	CurrentHash []byte
	DataBase    *badger.DB
	Lock        *sync.Mutex
}

// Iterator create a block iterator
func (chain *Chain) Iterator() *BlockIterator {
	iterator := BlockIterator{
		CurrentHash: chain.Tip,
		DataBase:    chain.DataBase,
		Lock:        &chain.Lock,
	}
	return &iterator
}

// Next get previous block by preHash
func (iterator *BlockIterator) Next() *Block {
	// get current block serialize data
	iterator.Lock.Lock()
	serializeData, err := ReadFromDB(iterator.DataBase, []byte(BlockTable), iterator.CurrentHash)
	iterator.Lock.Unlock()
	utils.HandleError(err)
	// deserialize data
	var currentBlock Block
	err = utils.Deserialize(serializeData, &currentBlock)
	utils.HandleError(err)
	// update next point
	iterator.CurrentHash = currentBlock.Header.PrevHash

	return &currentBlock
}

// 从区块链中找到高度在范围内的所有区块
func (chain *Chain) FindBlocksInRange(min, max uint64) []*Block {
	var blocksInRange []*Block

	iter := chain.Iterator()
	for {
		block := iter.Next()

		if block.Header.Height >= min && block.Header.Height <= max {
			// 如果区块高度在指定范围内，则将其加入结果列表
			blocksInRange = append(blocksInRange, block)
		}

		// 如果当前区块是创世区块，则停止遍历
		if block.IsGenesisBlock() {
			break
		}
	}

	return blocksInRange
}

// CreateChain create a new chain
// address use for genesis block
func CreateChain(address []byte, path, logPath string) (*Chain, error) {
	//initialize log
	l := utils.NewLogger("[chain] ", logPath)

	// create a database
	db, err := OpenDatabase(path)
	if err != nil {
		l.Panic("fail to create database")
		return nil, err
	}

	// Genesis node create GenesisBlock
	genesisBlock, err := NewGenesisBlock(address)
	if err != nil {
		l.Panic("fail to create genesis Block")
		return nil, err
	}

	// add latest block flag
	err = WriteToDB(db, []byte(BlockTable), []byte(TipHashKey), genesisBlock.Header.Hash)
	if err != nil {
		l.Panic("fail to write block data into database")
		return nil, err
	}

	chain := &Chain{
		Tip:        genesisBlock.Header.Hash,
		BestHeight: 1,
		DataBase:   db,
		log:        l,
	}
	// add genesis block
	chain.AddGenesisBlock(genesisBlock)

	return chain, nil
}

// LoadChain initialize chain from database
func LoadChain(path, logPath string) (*Chain, error) {
	// initialize log
	l := utils.NewLogger("[chain] ", logPath)

	// create a database
	db, err := OpenDatabase(path)
	if err != nil {
		l.Panic("fail to create database")
		return nil, err
	}

	// read tip hash from database
	latestHash, err := ReadFromDB(db, []byte(BlockTable), []byte(TipHashKey))

	// empty chain
	var chain *Chain
	if err != nil {
		//l.Println("Initialize a new chain")
		chain = &Chain{
			Tip:        nil,
			BestHeight: 0,
			DataBase:   db,
			log:        l,
		}
		return chain, nil
	}

	// get best height
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

// AddGenesisBlock add genesis block to chain and update UTXO set
func (chain *Chain) AddGenesisBlock(genesisBlock *Block) {
	// add genesis block to Chain
	data, err := utils.Serialize(genesisBlock)
	utils.HandleError(err)
	chain.Lock.Lock()
	err = WriteToDB(chain.DataBase, []byte(BlockTable), genesisBlock.Header.Hash, data)
	chain.Lock.Unlock()
	utils.HandleError(err)

	// add GenesisBlock output to UTXO set
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

// AddBlock add a block to chain
func (chain *Chain) AddBlock(block *Block) bool {
	// get current block previous hash
	preHash := block.Header.PrevHash

	// new block is genesis block
	if block.IsGenesisBlock() {
		// add block to chain
		chain.AddGenesisBlock(block)
		// update chain metadata
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

func (chain *Chain) GetTip() string {
	chain.Lock.Lock()
	defer chain.Lock.Unlock()
	return hex.EncodeToString(chain.Tip)
}

func (chain *Chain) GetHeight() uint64 {
	chain.Lock.Lock()
	defer chain.Lock.Unlock()
	return chain.BestHeight
}
