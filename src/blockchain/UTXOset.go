package blockchain

import (
	"BlockChain/src/utils"
	"badger"
	"encoding/hex"
)

// UTXO type
type UTXO struct {
	Index  int
	Output TXoutput
}

//type UTXOSet struct {
//	UTXODb *badger.DB
//}

// FindEnoughUTXOFromSet find enough UTXO from UTXO set
// return UTXO total value
// map[string][]int: TxID->[unspent output index]
func FindEnoughUTXOFromSet(utxoDb *badger.DB, address []byte, amount int) (int, map[string][]int) {
	utxos := make(map[string][]int)
	sum := 0
	err := utxoDb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		// traverse UTXO set
		prefix := []byte(ChainStateTable)
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()

			// get TxID
			key := item.KeyCopy(nil)[1:]
			id := hex.EncodeToString(key)
			err := item.Value(func(val []byte) error {
				// Deserialize value to []UTXO
				var utxo []UTXO
				err := utils.Deserialize(val, &utxo)
				if err != nil {
					return err
				}
				// get UTXO from []UTXO
				for _, out := range utxo {
					// belong to address
					if out.Output.CanBeUnlocked(address) {
						sum += out.Output.Value
						// add index
						utxos[id] = append(utxos[id], out.Index)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			if sum >= amount {
				break
			}
		}

		return nil
	})
	utils.HandleError(err)
	if sum >= amount {
		return sum, utxos
	} else {
		return 0, nil
	}
}

// GetBalanceFromSet calculate balance of address
func GetBalanceFromSet(utxoDb *badger.DB, address []byte) int {
	balance := 0
	err := utxoDb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		prefix := []byte(ChainStateTable)
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()

			err := item.Value(func(val []byte) error {
				// Deserialize value to []UTXO
				var utxo []UTXO
				err := utils.Deserialize(val, &utxo)
				if err != nil {
					return err
				}
				// get UTXO from []UTXO
				for _, out := range utxo {
					// belong to address
					if out.Output.CanBeUnlocked(address) {
						balance += out.Output.Value
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	utils.HandleError(err)

	return balance
}

// ReindexUTXOSet update UTXO set
// utxoMap from chain.FindUTXO
func ReindexUTXOSet(utxoDb *badger.DB, utxoMap map[string][]UTXO) {
	// delete all UTXO item from database
	err := utxoDb.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		prefix := []byte("c")
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			key := item.KeyCopy(nil)

			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		// add new UTXO item to database
		for id, utxos := range utxoMap {
			txId, err := hex.DecodeString(id)
			utils.HandleError(err)
			serializeData, err := utils.Serialize(utxos)
			utils.HandleError(err)
			err = txn.Set(append([]byte(ChainStateTable), txId...), serializeData)
			utils.HandleError(err)
		}
		return nil
	})
	utils.HandleError(err)
}

// UpdateUTXOSet update UTXO set when a new block add to chain
func UpdateUTXOSet(utxoDb *badger.DB, block *Block) bool {
	for _, tx := range block.Transactions {
		utxoMap := make(map[string][]UTXO)
		var utxos []UTXO
		// delete spent output
		for _, in := range tx.Inputs {
			id := hex.EncodeToString(in.TxID)
			if _, ok := utxoMap[id]; !ok {
				serializeData, err := ReadFromDB(utxoDb, []byte(ChainStateTable), in.TxID)
				if err != nil {
					return false
				}
				err = utils.Deserialize(serializeData, &utxos)
				if err != nil {
					return false
				}
				utxoMap[id] = utxos
			}
			var newUTXOs []UTXO
			for _, utxo := range utxoMap[id] {
				// delete input corresponding output
				if in.Index == utxo.Index {
					continue
				}
				newUTXOs = append(newUTXOs, utxo)
			}
			utxoMap[id] = newUTXOs
		}
		// add new UTXO from output
		var newUTXOs []UTXO
		for i, out := range tx.Outputs {
			newUTXO := UTXO{
				Index:  i,
				Output: out,
			}
			newUTXOs = append(newUTXOs, newUTXO)
		}
		utxoMap[hex.EncodeToString(tx.ID)] = newUTXOs

		// write new UTXOs into database
		for id, utxo := range utxoMap {
			txID, err := hex.DecodeString(id)
			if err != nil {
				return false
			}
			// utxo is empty
			if utxo == nil {
				err = DeleteFromDB(utxoDb, []byte(ChainStateTable), txID)
				continue
			}
			serializeData, err := utils.Serialize(utxo)
			if err != nil {
				return false
			}
			err = UpdateInDB(utxoDb, []byte(ChainStateTable), txID, serializeData)
			if err != nil {
				return false
			}
		}
	}
	return true
}
