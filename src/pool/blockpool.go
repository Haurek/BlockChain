package pool

import (
	"BlockChain/src/blockchain"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/state"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type BlockPool struct {
	mu             sync.Mutex
	pool           map[string]*blockchain.Block
	chain          *blockchain.Chain
	network        *p2pnet.P2PNet
	ws             *state.WorldState
	peerBestHeight uint64
	bestPeerID     string
	newBlock       chan *BlockMessage
	syncMsg        chan *BlockMessage
}

// NewBlockPool create a new block pool
func NewBlockPool(net *p2pnet.P2PNet, chain *blockchain.Chain, s *state.WorldState) (*BlockPool, error) {
	if net != nil {
		return nil, errors.New("unknown network")
	}

	pool := &BlockPool{
		pool:           make(map[string]*blockchain.Block),
		network:        net,
		chain:          chain,
		ws:             s,
		peerBestHeight: 0,
		newBlock:       make(chan *BlockMessage),
		syncMsg:        make(chan *BlockMessage),
	}
	return pool, nil
}

func (bp *BlockPool) Run() {
	// register receive callback func
	bp.network.RegisterCallback(p2pnet.BlockMsg, bp.OnReceive)

	// run block sync
	go bp.BlockSynchronization()

	// routine wait new block
	for {
	}

}

// OnReceive handle message receive from peer
func (bp *BlockPool) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.BlockMsg {
		return
	}
	var blockMsg BlockMessage
	err := json.Unmarshal(msgBytes, &blockMsg)
	if err != nil {
		return
	}
	// handle message receive
	switch blockMsg.Type {
	case SyncRequestMsg:
	case SyncResponseMsg:
	case BlockRequestMsg:
	case BlockResponseMsg:
		bp.syncMsg <- &blockMsg
	case NewBlockBroadcastMsg:
		bp.newBlock <- &blockMsg
	default:
		return
	}
}

func (bp *BlockPool) BlockSynchronization() {
	// broadcast current chain height
	syncReqMsg := SyncRequestMessage{
		NodeID:      bp.ws.SelfID,
		BlockHeight: bp.chain.BestHeight,
	}
	blockMsg := BlockMessage{
		Type: SyncRequestMsg,
		Data: syncReqMsg,
	}
	data, err := json.Marshal(blockMsg)
	if err != nil {
		return
	}
	p2pMsg := p2pnet.Message{
		Type: p2pnet.BlockMsg,
		Data: data,
	}
	bp.network.Broadcast(&p2pMsg)
	syncTimer := time.NewTimer(10 * time.Second)
	// run synchronization routine
	for {
		select {
		case syncMsg := <-bp.syncMsg:
			msg, t := syncMsg.SplitMessage()
			switch t {
			case SyncRequestMsg:
				// receive other peer sync request
				if request, ok := msg.(SyncRequestMessage); ok {
					// check best height
					// send sync response message
					syncResMsg := SyncResponseMessage{
						FromID:     bp.ws.SelfID,
						ToID:       request.NodeID,
						BestHeight: bp.chain.BestHeight,
					}
					blockMsg = BlockMessage{
						Type: SyncResponseMsg,
						Data: syncResMsg,
					}
					data, err = json.Marshal(blockMsg)
					if err != nil {
						return
					}
					p2pMsg = p2pnet.Message{
						Type: p2pnet.BlockMsg,
						Data: data,
					}
					// send response to node
					bp.network.BroadcastToPeer(&p2pMsg, request.NodeID)
				}
			case SyncResponseMsg:
				// receive other peer response
				if response, ok := msg.(SyncResponseMessage); ok {
					// check best height and update
					if response.BestHeight > bp.peerBestHeight {
						bp.UpdatePeerBestHeight(response.BestHeight, response.FromID)
					}
				}
			case BlockRequestMsg:
				// handle block request
				// 从区块链数据库取出请求的区块
				// TODO
				// 发送 BlockResponseMessage 消息回复请求方
				// TODO

			case BlockResponseMsg:
				// handle block response
				// 取出BlockResponseMessage中的区块信息
				// TODO
				// 判断收到的区块是否可直接加入区块链，可以则直接插入
				// TODO
				// 插入后遍历区块池判断是否可以继续插入区块
				bp.Reindex()
				// 无法插入的区块链则暂时加入区块池BlockPool
				// TODO
			default:

			}
		case <-syncTimer.C:
			// 计时器判断同步请求是否完成
			// 请求完成，开始向拥有最长链的节点发送BlockRequestMessage消息请求区块
			// TODO
		}
	}
}

// AddBlock add block to pool
func (bp *BlockPool) AddBlock(block *blockchain.Block) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	id := hex.EncodeToString(block.Header.Hash)
	if _, exists := bp.pool[id]; !exists {
		bp.pool[id] = block
	}
}

func (bp *BlockPool) Reindex() {
	// TODO
}

// GetBlock get block from pool by hash
func (bp *BlockPool) GetBlock(hash []byte) *blockchain.Block {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	id := hex.EncodeToString(hash)
	if _, exists := bp.pool[id]; !exists {
		return bp.pool[id]
	}
	return nil
}

// RemoveBlock remove block from pool by hash
func (bp *BlockPool) RemoveBlock(hash []byte) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	id := hex.EncodeToString(hash)
	if _, exists := bp.pool[id]; !exists {
		delete(bp.pool, id)
	}
}

// HaveBlock check a block in pool
func (bp *BlockPool) HaveBlock(hash []byte) bool {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	id := hex.EncodeToString(hash)
	_, exists := bp.pool[id]
	return exists
}

// Count block num in pool
func (bp *BlockPool) Count() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	return len(bp.pool)
}

func (bp *BlockPool) UpdatePeerBestHeight(height uint64, id string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.peerBestHeight = height
	bp.bestPeerID = id
}
