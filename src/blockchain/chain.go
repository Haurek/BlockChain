package blockchain

import (
	"BlockChain/src/utils"
	"badger"
	"bytes"
	"encoding/hex"
	"errors"
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
	utils.HandleError(err)
	// deserialize data
	var currentBlock Block
	err = utils.Deserialize(serializeData, &currentBlock)
	utils.HandleError(err)
	// update next point
	iterator.CurrentHash = currentBlock.Header.PrevHash

	return &currentBlock
}

// CreateChain create a new chain
// address use for genesis block
func CreateChain(address []byte) *Chain {
	// create a database
	db, err := OpenDatabase(DataBasePath)
	if err != nil {
		return nil
	}
	// Genesis node create GenesisBlock
	genesisBlock, err := NewGenesisBlock(address)
	utils.HandleError(err)

	// add latest block flag
	err = WriteToDB(db, []byte(BlockTable), []byte(TipHashKey), genesisBlock.Header.Hash)
	utils.HandleError(err)

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
func LoadChain(path string) (*Chain, error) {
	//// check chain database exist
	//if _, err := os.Stat(path); os.IsNotExist(err) {
	//	fmt.Println("No Chain found")
	//	return nil
	//}
	//// load local database
	//opts := badger.DefaultOptions(path)
	//opts.Logger = nil
	//db, err := badger.Open(opts)
	//utils.HandleError(err)
	db, err := OpenDatabase(path)
	if err != nil {
		return nil, err
	}
	// read tip hash from database
	latestHash, err := ReadFromDB(db, []byte(BlockTable), []byte(TipHashKey))
	utils.HandleError(err)

	chain := &Chain{
		Tip:      latestHash,
		DataBase: db,
	}

	return chain, nil
}

// AddGenesisBlock add genesis block to chain and update UTXO set
func (chain *Chain) AddGenesisBlock(genesisBlock *Block) {
	// add genesis block to Chain
	data, err := utils.Serialize(genesisBlock)
	utils.HandleError(err)
	err = WriteToDB(chain.DataBase, []byte(BlockTable), genesisBlock.Header.Hash, data)
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
	err = WriteToDB(chain.DataBase, []byte(ChainStateTable), genesisBlock.Transactions[0].ID, data)
	utils.HandleError(err)
}

// AddBlock add a block to chain
func (chain *Chain) AddBlock(block *Block) bool {
	// get current block previous hash
	preHash := block.Header.PrevHash

	// get previous block from database
	preBlockData, err := ReadFromDB(chain.DataBase, []byte(BlockTable), preHash)
	utils.HandleError(err)
	var preBlock Block
	err = utils.Deserialize(preBlockData, &preBlock)
	utils.HandleError(err)

	// check hash and PoW verify
	if bytes.Equal(preBlock.Header.Hash, block.Header.PrevHash) {
		// add to database
		serializeData, err := utils.Serialize(block)
		utils.HandleError(err)
		err = WriteToDB(chain.DataBase, []byte(BlockTable), block.Header.Hash, serializeData)
		utils.HandleError(err)

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
	utils.HandleError(err)
	var block Block
	err = utils.Deserialize(serializeData, &block)
	utils.HandleError(err)
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
