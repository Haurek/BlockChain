package pool

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/network"
	"BlockChain/src/utils"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
)

type TxPool struct {
	full       int                                // signal primary when TxPool is full
	pool       map[string]*blockchain.Transaction // Tx cache
	network    *p2pnet.P2PNet                     // p2p network
	FullSignal chan *struct{}                     // full chan
	log        *log.Logger
	lock       sync.Mutex
}

// NewTxPool create a new transaction pool
func NewTxPool(f int, net *p2pnet.P2PNet, logPath string) *TxPool {
	// initialize logger
	l := utils.NewLogger("[TxPool] ", logPath)

	if net == nil {
		l.Panic("unknown network")
	}

	pool := &TxPool{
		full:       f,
		pool:       make(map[string]*blockchain.Transaction),
		network:    net,
		log:        l,
		FullSignal: make(chan *struct{}),
	}
	return pool
}

func (tp *TxPool) Run() {
	// register receive callback func
	tp.log.Println("Run Transaction Pool")
	tp.network.RegisterCallback(p2pnet.TransactionMsg, tp.OnReceive)
}

// AddTransaction add transaction to pool
func (tp *TxPool) AddTransaction(tx *blockchain.Transaction) {
	id := hex.EncodeToString(tx.ID[:])
	tp.lock.Lock()
	if _, exists := tp.pool[id]; !exists {
		tp.log.Println("Add Transaction to pool: ", id)
		tp.pool[id] = tx
	}
	if len(tp.pool) >= tp.full {
		// TxPool full, interrupt consensus layer
		tp.FullSignal <- &struct{}{}
	}
	tp.lock.Unlock()
}

// GetTransactions get all transactions from pool
func (tp *TxPool) GetTransactions() map[string]*blockchain.Transaction {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	transactions := make(map[string]*blockchain.Transaction)
	for id, tx := range tp.pool {
		transactions[id] = tx
	}

	return transactions
}

// RemoveTransaction remove transaction from pool by ID
func (tp *TxPool) RemoveTransaction(id string) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	if _, exists := tp.pool[id]; exists {
		delete(tp.pool, id)
		tp.log.Println("Delete Transaction from pool: ", id)
	}
}

func (tp *TxPool) HaveTransaction(id string) bool {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	_, exists := tp.pool[id]
	return exists
}

// Count transaction num in pool
func (tp *TxPool) Count() int {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	return len(tp.pool)
}

// OnReceive handle transaction message receive from peer
func (tp *TxPool) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.TransactionMsg {
		return
	}
	var txMsg TxMessage
	tp.log.Println("Receive new tx")
	err := json.Unmarshal(msgBytes, &txMsg)
	if err != nil {
		tp.log.Println("Unmarshal message fail")
		return
	}
	switch txMsg.Type {
	case SendTxMsg:
		var tx blockchain.Transaction
		err = json.Unmarshal(txMsg.TxBytes, &tx)
		if err != nil {
			tp.log.Println("Unmarshal tx fail")
			return
		}

		// add tx to pool
		if tp.HaveTransaction(hex.EncodeToString(tx.ID[:])) {
			tp.log.Println("Transaction already in pool")
			return
		}
		tp.AddTransaction(&tx)
		tp.log.Println("Add Transaction to pool: ", hex.EncodeToString(tx.ID[:]))

		// broadcast to other peers
		msg := &p2pnet.Message{
			Type: p2pnet.TransactionMsg,
			Data: msgBytes,
		}
		tp.log.Println("Broadcast Transaction to peers: ", hex.EncodeToString(tx.ID[:]))
		tp.network.BroadcastExceptPeer(msg, peerID)
	default:
		return
	}
}
