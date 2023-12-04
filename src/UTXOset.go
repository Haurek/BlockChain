package main

import (
	"badger"
	"encoding/hex"
	"errors"
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
				err := Deserialize(val, &utxo)
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
	HandleError(err)
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
				err := Deserialize(val, &utxo)
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
	HandleError(err)

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
			HandleError(err)
			serializeData, err := Serialize(utxos)
			HandleError(err)
			err = txn.Set(append([]byte(ChainStateTable), txId...), serializeData)
			HandleError(err)
		}
		return nil
	})
	HandleError(err)
}

// UpdateUTXOSet update UTXO set when a new block add to chain
func UpdateUTXOSet(utxoDb *badger.DB, block *Block) {
	for _, tx := range block.Transactions {
		if tx.IsCoinBase() == false {
			var utxos []UTXO
			var newUTXO []UTXO
			serializeData, err := ReadFromDB(utxoDb, []byte(ChainStateTable), tx.ID)
			// not found in set
			// add all output to database
			if errors.Is(err, badger.ErrKeyNotFound) {
				var item []UTXO
				for i, out := range tx.Outputs {
					utxo := UTXO{
						Index:  i,
						Output: out,
					}
					item = append(item, utxo)
				}
				serializeData, err = Serialize(&item)
				HandleError(err)
				err = WriteToDB(utxoDb, []byte(ChainStateTable), tx.ID, serializeData)
				HandleError(err)
			} else {
				err = Deserialize(serializeData, &utxos)

				for _, utxo := range utxos {
					// find input corresponding output
					for _, in := range tx.Inputs {
						if in.Index == utxo.Index {
							break
						}
						newUTXO = append(newUTXO, utxo)
					}
				}

				serializeData, err = Serialize(newUTXO)
				HandleError(err)
				err = UpdateInDB(utxoDb, []byte(ChainStateTable), tx.ID, serializeData)
				HandleError(err)
			}
		}
	}
}
