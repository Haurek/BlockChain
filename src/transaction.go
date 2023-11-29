package BlockChain

import (
	"crypto/sha256"
)

// Transaction include inputs and outputs
type Transaction struct {
	ID     [32]byte
	Inputs []*TXinput
	Output []*TXoutput
}

//func (Tx *Transaction) NewTransaction(wallet *Wallet, to []byte, amount int) *Transaction {
//	// Find enough output from wallet address
//	// maybe search from database
//	// TODO
//	//balance, walletOutputs := FindEnoughUnspentOutput()
//	if balance < amount {
//		fmt.Println("Not enough balance")
//		return nil
//	}
//
//	// build inputs
//	var inputs []*TXinput
//	for i, out := range walletOutputs {
//		input := &TXinput{}
//		input.NewTXinput(i, out.Value, wallet)
//		inputs = append(inputs, input)
//	}
//	// sign inputs
//	Tx.Inputs = inputs
//	Tx.SignTransaction(wallet)
//
//	// build outputs
//	var outputs []*TXoutput
//	output := &TXoutput{}
//	output.NewTXoutput(0, amount, to)
//	outputs = append(outputs, output)
//	// charge
//	if balance > amount {
//		charge := &TXoutput{}
//		charge.NewTXoutput(1, balance-amount, wallet.Address)
//		outputs = append(outputs, charge)
//	}
//	Tx.Output = outputs
//
//	// generate Transaction hash as ID
//	Tx.ID = Tx.Hash()
//	return Tx
//}
//
//// SignTransaction sign each input of transaction
//func (Tx *Transaction) SignTransaction(wallet *Wallet) {
//	for i, input := range Tx.Inputs {
//		// get previous transaction of a input
//		// maybe get from chain(database)
//		// TODO
//		//preTx := GetPreTransaction(input)
//		message := preTx.Serialize()
//		signature, err := wallet.Sign(message)
//		if err != nil {
//			fmt.Println("fail to generate signature")
//		}
//		Tx.Inputs[i].Signature = signature
//	}
//
//}
//
//// VerifyTransaction verify each input of transaction
//func VerifyTransaction(Tx *Transaction, publicKey []byte) bool {
//	// coinbase
//	// TODO
//
//	for _, input := range Tx.Inputs {
//		// TODO
//		// verify format
//
//		// TODO
//		// verify signature
//		signature := input.Signature
//		//preTx := GetPreTransaction(input)
//		message := preTx.Serialize()
//		if Verify(publicKey, message, signature) == false {
//			return false
//		}
//		// TODO
//		// avoid double spent locally
//	}
//	return true
//}

// Hash generate Transaction hash
func (Tx *Transaction) Hash() []byte {
	raw := Tx.Serialize()
	var hash [32]byte
	hash = sha256.Sum256(raw)
	return hash[:]
}

// Serialize Transaction struct
func (Tx *Transaction) Serialize() []byte {
	// TODO
	return nil
}

// Deserialize []byte data to Transaction type
func (Tx *Transaction) Deserialize(raw []byte) *Transaction {
	// TODO
	return nil
}
