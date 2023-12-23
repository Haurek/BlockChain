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
	// utxos stores unspent output indexes for each transaction ID
	utxos := make(map[string][]int)

	// sum holds the accumulated value of unspent outputs
	sum := 0

	// Read-only transaction to traverse the UTXO set
	err := utxoDb.View(func(txn *badger.Txn) error {
		// Iterator options to enhance performance
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		// Traverse the UTXO set starting from the specified prefix
		prefix := []byte(ChainStateTable)
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()

			// Extract the transaction ID
			key := item.KeyCopy(nil)[1:] // Assuming the first byte is skipped for the prefix
			id := hex.EncodeToString(key)

			// Process the value associated with the key (i.e., the serialized UTXO)
			err := item.Value(func(val []byte) error {
				// Deserialize the value to []UTXO
				var utxo []UTXO
				err := utils.Deserialize(val, &utxo)
				if err != nil {
					return err
				}

				// Iterate through the UTXOs
				for _, out := range utxo {
					// Check if the output belongs to the given address
					if out.Output.CanBeUnlocked(address) {
						// Accumulate the value of unspent outputs
						sum += out.Output.Value
						// Store the index of the unspent output for the corresponding transaction ID
						utxos[id] = append(utxos[id], out.Index)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}

			// Exit the loop if the accumulated sum meets or exceeds the required amount
			if sum >= amount {
				break
			}
		}
		return nil
	})
	utils.HandleError(err)

	// Return the total accumulated sum of unspent outputs and their corresponding indexes
	if sum >= amount {
		return sum, utxos
	} else {
		return 0, nil // Return zero if there's not enough unspent value
	}
}

// GetBalanceFromSet calculates the balance of an address by summing up the values of its unspent outputs
func GetBalanceFromSet(utxoDb *badger.DB, address []byte) int {
	// Initialize the balance to zero
	balance := 0

	// Read-only transaction to traverse the UTXO set
	err := utxoDb.View(func(txn *badger.Txn) error {
		// Iterator options to enhance performance
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		// Seek to the prefix associated with the UTXO set
		prefix := []byte(ChainStateTable)
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()

			// Process the value associated with the key (i.e., the serialized UTXO)
			err := item.Value(func(val []byte) error {
				// Deserialize the value to []UTXO
				var utxo []UTXO
				err := utils.Deserialize(val, &utxo)
				if err != nil {
					return err
				}

				// Iterate through the UTXOs
				for _, out := range utxo {
					// Check if the output belongs to the given address
					if out.Output.CanBeUnlocked(address) {
						// Accumulate the value of unspent outputs
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

	// Return the calculated balance for the given address
	return balance
}

// ReindexUTXOSet updates the UTXO set in the database with a new set of UTXOs
// utxoMap represents the updated UTXO set obtained from chain.FindUTXO
func ReindexUTXOSet(utxoDb *badger.DB, utxoMap map[string][]UTXO) {
	// Delete all existing UTXO items from the database
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

		// Add the new UTXO items to the database
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

// UpdateUTXOSet updates the UTXO set when a new block is added to the blockchain
func UpdateUTXOSet(utxoDb *badger.DB, block *Block) bool {
	for _, tx := range block.Transactions {
		utxoMap := make(map[string][]UTXO)
		var utxos []UTXO

		// Delete spent outputs and update UTXO map
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
				// Delete the input corresponding to the output
				if in.Index == utxo.Index {
					continue
				}
				newUTXOs = append(newUTXOs, utxo)
			}
			utxoMap[id] = newUTXOs
		}

		// Add new UTXOs from transaction outputs to the UTXO map
		var newUTXOs []UTXO
		for i, out := range tx.Outputs {
			newUTXO := UTXO{
				Index:  i,
				Output: out,
			}
			newUTXOs = append(newUTXOs, newUTXO)
		}
		utxoMap[hex.EncodeToString(tx.ID)] = newUTXOs

		// Write the updated UTXOs into the database
		for id, utxo := range utxoMap {
			txID, err := hex.DecodeString(id)
			if err != nil {
				return false
			}
			// If the UTXO is empty, delete it from the database
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
