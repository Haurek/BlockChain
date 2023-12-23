package blockchain

import (
	"BlockChain/src/mycrypto"
	"BlockChain/src/utils"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// Transaction include inputs and outputs
type Transaction struct {
	ID      []byte     `json:"ID"`
	Inputs  []TXinput  `json:"Inputs"`
	Outputs []TXoutput `json:"Outputs"`
}

// IsCoinBase check coinbase transaction
func (Tx *Transaction) IsCoinBase() bool {
	return len(Tx.Inputs) == 1 && Tx.Inputs[0].Index == -1
}

// NewCoinbaseTx create coinbase transaction
func NewCoinbaseTx(to []byte, reward int) *Transaction {
	input := TXinput{
		TxID:           []byte{},
		Index:          -1,
		FromAddress:    []byte{},
		Signature:      []byte{},
		PublicKeyBytes: []byte{},
	}
	output := NewTXoutput(reward, to)
	Tx := &Transaction{
		ID:      nil,
		Inputs:  []TXinput{input},
		Outputs: []TXoutput{*output},
	}
	Tx.ID = HashTransaction(Tx)
	return Tx
}

// NewTransaction create new transaction
func NewTransaction(wallet *Wallet, chain *Chain, to []byte, amount int) (*Transaction, error) {
	// check valid address
	if !CheckAddress(wallet.address) || !CheckAddress(to) {
		return nil, errors.New("wrong address")
	}
	// check valid amount
	if amount <= 0 {
		return nil, errors.New("wrong amount")
	}

	// Find enough UTXO from UTXO set by wallet address
	balance, utxosMap := FindEnoughUTXOFromSet(chain.DataBase, wallet.address, amount)
	if balance < amount {
		return nil, errors.New("Not enough balance")
	}

	// build inputs
	var inputs []TXinput
	for id, indexSet := range utxosMap {
		// decode id
		txId, err := hex.DecodeString(id)
		utils.HandleError(err)
		for _, index := range indexSet {
			input := NewTXinput(index, wallet.address, txId, wallet.GetPublicKeyBytes())
			preTx, err := chain.FindTransaction(input.TxID)
			if err != nil {
				return nil, errors.New("fail to find previous Tx")
			}
			// sign input
			message := HashTransaction(preTx)
			sig, err := mycrypto.Sign(wallet.privateKey, message)
			utils.HandleError(err)
			input.SetSignature(sig)
			inputs = append(inputs, *input)
		}
	}

	// build outputs
	var outputs []TXoutput
	output := NewTXoutput(amount, to)
	outputs = append(outputs, *output)

	// change
	if balance > amount {
		change := NewTXoutput(balance-amount, wallet.GetAddress())
		outputs = append(outputs, *change)
	}
	Tx := &Transaction{
		ID:      nil,
		Inputs:  inputs,
		Outputs: outputs,
	}
	Tx.ID = HashTransaction(Tx)
	return Tx, nil
}

// VerifyTransaction verify each input of transaction before pack a block
func VerifyTransaction(chain *Chain, Tx *Transaction) bool {
	// coinbase
	if Tx.IsCoinBase() == true {
		return true
	}

	for _, input := range Tx.Inputs {
		// find previous Tx of input
		preTx, err := chain.FindTransaction(input.TxID)
		if err != nil {
			fmt.Println("fail to find previous Tx")
			return false
		}

		// verify signature
		signature := input.Signature
		message := HashTransaction(preTx)
		if mycrypto.Verify(mycrypto.Bytes2PublicKey(input.PublicKeyBytes), message, signature) == false {
			return false
		}
	}
	return true
}

// HashTransaction hash of all inputs and outputs in Tx
func HashTransaction(tx *Transaction) []byte {
	txCopy := tx.TrimmedCopy()
	txCopy.ID = []byte{}
	raw, err := json.Marshal(txCopy)
	if err != nil {
		fmt.Println("Error during serialization:", err)
		return nil
	}

	return utils.Sha256Hash(raw)
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXinput
	var outputs []TXoutput

	for _, vin := range tx.Inputs {
		inputs = append(inputs, TXinput{vin.TxID, vin.Index, vin.FromAddress, vin.Signature, vin.PublicKeyBytes})
	}

	for _, vout := range tx.Outputs {
		outputs = append(outputs, TXoutput{vout.Value, vout.ToAddress, vout.PublicKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}
