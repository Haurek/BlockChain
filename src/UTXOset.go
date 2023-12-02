package BlockChain

import (
	badger "badger-4.2.0"
)

// UTXO type
type UTXO struct {
	outputs []TXoutput
}

//// NewUTXO create new UTXO
//func NewUTXO(id []byte, index, value int) *UTXO {
//	return &UTXO{
//		TxID:  id,
//		Index: index,
//		Value: value,
//	}
//}

type UTXOSet struct {
	UTXODb *badger.DB
}

// FindEnoughUTXO find enough UTXO from UTXO set
func (u *UTXOSet) FindEnoughUTXO(address []byte, amount int) (int, []*TXoutput) {
	utxos := make([]*TXoutput, MaxUTXOSize)
	sum := 0
	err := u.UTXODb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		prefix := []byte(ChainStateTable)
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()

			err := item.Value(func(val []byte) error {
				var utxo *UTXO
				err := Deserialize(val, utxo)
				if err != nil {
					return err
				}
				if out.CanBeUnlocked(address) {
					sum += out.Value
					utxos = append(utxos, out)
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

// FindUTXO find all UTXO of address
func (u *UTXOSet) FindUTXO(address []byte) (int, []*TXoutput) {
	utxos := make([]*TXoutput, MaxUTXOSize)
	balance := 0
	err := u.UTXODb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		prefix := []byte(ChainStateTable)
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()

			err := item.Value(func(val []byte) error {
				var out *TXoutput
				err := Deserialize(val, out)
				if err != nil {
					return err
				}
				if out.CanBeUnlocked(address) {
					balance += out.Value
					utxos = append(utxos, out)
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

	return balance, utxos
}
