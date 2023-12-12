package blockchain

import (
	"BlockChain/src/mycrypto"
	"BlockChain/src/utils"
	"encoding/hex"
	"errors"
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
	if !IsValidAddress(wallet.address) || !IsValidAddress(to) {
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
		for index := range indexSet {
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
	txCopy := *tx
	txCopy.ID = []byte{}
	raw, err := utils.Serialize(txCopy)
	if err != nil {
		fmt.Println("Error during serialization:", err)
		return nil
	}

	return utils.Sha256Hash(raw)
}

//// 定义交易信息池结构
//type TxPool struct {
//	Transactions []*Transaction
//}
//
//// 创建一个新的交易信息池
//func NewTxPool() *TxPool {
//	return &TxPool{
//		Transactions: make([]*Transaction, MaxTxPoolSize),
//	}
//}
//
//// 将新的交易添加到交易信息池中
//func (tp *TxPool) AddTransaction(tx *Transaction) {
//	tp.Transactions = append(tp.Transactions, tx)
//}
//
//// 获取当前交易信息池中的所有交易
//func (tp *TxPool) GetTransactions() []*Transaction {
//	return tp.Transactions
//}
//
//// 清空交易信息池
//func (tp *TxPool) ClearPool() {
//	//clear(tp.Transactions)
//	tp.Transactions = make([]*Transaction, 0)
//}
