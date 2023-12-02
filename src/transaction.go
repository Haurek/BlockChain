package BlockChain

import (
	"encoding/hex"
	"fmt"
)

// Transaction include inputs and outputs
type Transaction struct {
	ID      []byte
	Inputs  []TXinput
	Outputs []TXoutput
}

// IsCoinBase check coinbase transaction
func (Tx *Transaction) IsCoinBase() bool {
	return len(Tx.Inputs) == 0 && Tx.Inputs[0].Index == -1
}

// NewCoinbaseTx create coinbase transaction
func NewCoinbaseTx(to []byte) *Transaction {
	input := TXinput{
		TxID:           nil,
		Index:          -1,
		FromAddress:    nil,
		Signature:      nil,
		PublicKeyBytes: nil,
	}
	output := NewTXoutput(Reward, to)
	Tx := &Transaction{
		ID:      nil,
		Inputs:  []TXinput{input},
		Outputs: []TXoutput{*output},
	}
	return Tx
}

// NewTransaction create new transaction
func NewTransaction(wallet *Wallet, chain *Chain, to []byte, amount int) *Transaction {
	// Find enough UTXO from wallet address
	balance, utxosMap := FindEnoughUTXOFromSet(chain.DataBase, wallet.address, amount)
	if balance < amount {
		fmt.Println("Not enough balance")
		return nil
	}

	// build inputs
	var inputs []TXinput
	for id, indexSet := range utxosMap {
		// decode id
		txId, err := hex.DecodeString(id)
		HandleError(err)
		for index := range indexSet {
			input := NewTXinput(index, wallet.address, txId, wallet.GetPublicKeyBytes())
			preTx, err := chain.FindTransaction(input.TxID)
			if err != nil {
				fmt.Println("fail to find previous Tx")
				return nil
			}
			// sign input
			message := HashTransaction(preTx)
			sig, err := Sign(wallet.privateKey, message)
			if err != nil {
				return nil
			}
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
	return Tx
}

// VerifyTransaction verify each input of transaction
func VerifyTransaction(chain *Chain, Tx *Transaction, publicKeyBytes []byte) bool {
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

		signature := input.Signature
		//preTx := GetPreTransaction(input)
		message := HashTransaction(preTx)
		if Verify(Bytes2PublicKey(input.PublicKeyBytes), message, signature) == false {
			return false
		}
		// TODO
		// avoid double spent locally
	}
	return true
}

// HashTransaction hash of all inputs and outputs in Tx
func HashTransaction(tx *Transaction) []byte {
	txCopy := *tx
	txCopy.ID = []byte{}
	raw, err := Serialize(txCopy)
	if err != nil {
		fmt.Println("Error during serialization:", err)
		return nil
	}

	return Sha256Hash(raw)
}

// 定义交易信息池结构
type TxPool struct {
	Transactions []*Transaction
}

// 创建一个新的交易信息池
func NewTxPool() *TxPool {
	return &TxPool{
		Transactions: make([]*Transaction, 0),
	}
}

// 将新的交易添加到交易信息池中
func (tp *TxPool) AddTransaction(tx *Transaction) {
	tp.Transactions = append(tp.Transactions, tx)
}

// 获取当前交易信息池中的所有交易
func (tp *TxPool) GetTransactions() []*Transaction {
	return tp.Transactions
}

// 清空交易信息池
func (tp *TxPool) ClearPool() {
	tp.Transactions = make([]*Transaction, 0)
}
