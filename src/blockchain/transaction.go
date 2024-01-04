package blockchain

import (
	"BlockChain/src/mycrypto"
	"BlockChain/src/utils"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// Transaction represents a collection of inputs and outputs in a transaction
type Transaction struct {
	ID      []byte     `json:"ID"`      // ID represents the unique identifier for the transaction
	Inputs  []TXinput  `json:"Inputs"`  // Inputs include the details of the transaction inputs
	Outputs []TXoutput `json:"Outputs"` // Outputs include the details of the transaction outputs
}

// IsCoinBase checks if the transaction is a coinbase transaction
func (Tx *Transaction) IsCoinBase() bool {
	return len(Tx.Inputs) == 1 && Tx.Inputs[0].Index == -1
}

// NewCoinbaseTx creates a coinbase transaction, used to reward miners
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

// NewTransaction creates a new transaction transferring a specified amount between wallets
func NewTransaction(wallet *Wallet, chain *Chain, to []byte, amount int) (*Transaction, error) {
	// Check the validity of the addresses and the amount
	if !CheckAddress(wallet.address) || !CheckAddress(to) {
		return nil, errors.New("wrong address")
	}
	if amount <= 0 {
		return nil, errors.New("wrong amount")
	}

	// Find enough Unspent Transaction Outputs (UTXOs) from the UTXO set by the wallet's address
	balance, utxosMap := FindEnoughUTXOFromSet(chain.DataBase, wallet.address, amount)
	if balance < amount {
		return nil, errors.New("Not enough balance")
	}

	var inputs []TXinput
	for id, indexSet := range utxosMap {
		// Decode the transaction ID
		txId, err := hex.DecodeString(id)
		utils.HandleError(err)
		for _, index := range indexSet {
			input := NewTXinput(index, wallet.address, txId, wallet.GetPublicKeyBytes())
			preTx, err := chain.FindTransaction(input.TxID)
			if err != nil {
				return nil, errors.New("fail to find previous Tx")
			}
			// Sign the input using the wallet's private key
			message := HashTransaction(preTx)
			sig, err := mycrypto.Sign(wallet.privateKey, message)
			utils.HandleError(err)
			input.SetSignature(sig)
			inputs = append(inputs, *input)
		}
	}

	var outputs []TXoutput
	output := NewTXoutput(amount, to)
	outputs = append(outputs, *output)

	// Calculate the change if the wallet has more balance than the transaction amount
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

// VerifyTransactions Verify Txs Signature and check double spent
func VerifyTransactions(chain *Chain, Txs []*Transaction) bool {
	usedOutputs := make(map[string]struct{})
	for _, Tx := range Txs {
		// Check if it's a coinbase transaction
		if Tx.IsCoinBase() == true {
			continue
		}

		for _, input := range Tx.Inputs {
			// Find the previous transaction for each input
			preTx, err := chain.FindTransaction(input.TxID)
			if err != nil {
				chain.log.Println("fail to find previous Tx")
				return false
			}
			// check double spent
			outputIdentifier := fmt.Sprintf("%s:%d", input.TxID, input.Index)
			if _, exists := usedOutputs[outputIdentifier]; exists {
				chain.log.Println("Double spent")
				return false
			}
			usedOutputs[outputIdentifier] = struct{}{}

			// Verify the signature for each input
			signature := input.Signature
			message := HashTransaction(preTx)
			if mycrypto.Verify(mycrypto.Bytes2PublicKey(input.PublicKeyBytes), message, signature) == false {
				chain.log.Println("Signature verify error")
				return false
			}
		}
	}
	return true
}

// HashTransaction computes the hash of a transaction using its inputs and outputs
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

// TrimmedCopy creates a copy of the transaction with trimmed inputs and outputs
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
