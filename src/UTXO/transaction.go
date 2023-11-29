package UTXO

import (
	"BlockChain/src/Wallet"
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

// Transaction include inputs and outputs
type Transaction struct {
	ID     []byte
	Inputs []TXinput
	Output []TXoutput
}

func (Tx *Transaction) NewTransaction(wallet *Wallet.Wallet, to []byte, amount int) *Transaction {
	// Find enough output from wallet address
	// maybe search from database
	// TODO
	//balance, walletOutputs := FindEnoughUnspentOutput()
	balance := 0
	walletOutputs := make([]TXoutput, 2)
	if balance < amount {
		fmt.Println("Not enough balance")
		return nil
	}

	// build inputs
	var inputs []TXinput
	for i, out := range walletOutputs {
		input := &TXinput{}
		input.NewTXinput(i, out.Value, wallet)
		inputs = append(inputs, *input)
	}

	// build outputs
	var outputs []TXoutput
	output := &TXoutput{}
	output.NewTXoutput(0, amount, to)
	outputs = append(outputs, *output)
	// charge
	if balance > amount {
		charge := &TXoutput{}
		charge.NewTXoutput(1, balance-amount, wallet.Address)
		outputs = append(outputs, *charge)
	}

	// sign inputs
	Tx.SignTransaction(wallet.PrivateKey)

	// generate Transaction hash
	Tx.ID = Tx.Hash()
	Tx.Inputs = inputs
	Tx.Output = outputs
	return Tx
}

// SignTransaction sign each input of transaction
func (Tx *Transaction) SignTransaction(privateKey *ecdsa.PrivateKey) {
	// TODO
}

// VerifyTransaction verify each output of transaction
func (Tx *Transaction) VerifyTransaction(wallet *Wallet.Wallet) {
	// TODO
}

// Hash generate Transaction hash
func (Tx *Transaction) Hash() [32]byte {
	raw := Tx.Serialize()
	var hash [32]byte
	hash = sha256.Sum256(raw)
	return hash
}

// Serialize Transaction struct
func (Tx *Transaction) Serialize() []byte {
	// TODO
}

// Deserialize []byte data to Transaction type
func (Tx *Transaction) Deserialize(raw []byte) *Transaction {
	// TODO
}
