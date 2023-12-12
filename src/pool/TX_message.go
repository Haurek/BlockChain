package pool

import (
	"BlockChain/src/blockchain"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/utils"
	"encoding/json"
)

// TxMsgType type of transaction message
type TxMsgType int32

const (
	SendTxMsg TxMsgType = 0x00
)

// TxMessage is type of transaction message
type TxMessage struct {
	Type    TxMsgType `json:"type"`
	TxBytes []byte    `json:"tx"`
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
