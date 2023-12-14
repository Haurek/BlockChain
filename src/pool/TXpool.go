package pool

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/network"
	"BlockChain/src/utils"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
)

type TxPool struct {
	mu      sync.Mutex
	pool    map[string]*blockchain.Transaction
	network *p2pnet.P2PNet
}

// NewTxPool create a new transaction pool
func NewTxPool(net *p2pnet.P2PNet) (*TxPool, error) {
	if net != nil {
		return nil, errors.New("unknown network")
	}

	pool := &TxPool{
		pool:    make(map[string]*blockchain.Transaction),
		network: net,
	}
	return pool, nil
}

func (tp *TxPool) Run() {
	// register receive callback func
	tp.network.RegisterCallback(p2pnet.TransactionMsg, tp.OnReceive)
}

// AddTransaction add transaction to pool
func (tp *TxPool) AddTransaction(tx *blockchain.Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	id := hex.EncodeToString(tx.ID[:])
	if _, exists := tp.pool[id]; !exists {
		tp.pool[id] = tx
		//fmt.Printf("Added transaction to pool: %v\n", tx)
	}
}

// GetTransactions get all transactions from pool
func (tp *TxPool) GetTransactions() map[string]*blockchain.Transaction {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	transactions := make(map[string]*blockchain.Transaction)
	for id, tx := range tp.pool {
		transactions[id] = tx
	}

	return transactions
}

// RemoveTransaction remove transaction from pool by ID
func (tp *TxPool) RemoveTransaction(id string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if _, exists := tp.pool[id]; !exists {
		delete(tp.pool, id)
		//fmt.Printf("Added transaction to pool: %v\n", tx)
	}
}

func (tp *TxPool) HaveTransaction(id string) bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	_, exists := tp.pool[id]
	return exists
}

// Count transaction num in pool
func (tp *TxPool) Count() int {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	return len(tp.pool)
}

// OnReceive handle transaction message receive from peer
func (tp *TxPool) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.TransactionMsg {
		return
	}
	var txMsg TxMessage
	err := json.Unmarshal(msgBytes, &txMsg)
	if err != nil {
		return
	}
	switch txMsg.Type {
	case SendTxMsg:
		var tx blockchain.Transaction
		err = utils.Deserialize(txMsg.TxBytes, &tx)
		if err != nil {
			return
		}
		// add tx to pool
		if tp.HaveTransaction(hex.EncodeToString(tx.ID[:])) {
			return
		}
		tp.AddTransaction(&tx)
		// broadcast to other peers
		msg := &p2pnet.Message{
			Type: p2pnet.TransactionMsg,
			Data: msgBytes,
		}
		tp.network.BroadcastExceptPeer(msg, peerID)
	default:
		return
	}
}
