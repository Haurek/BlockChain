package pool

import (
	"BlockChain/src/blockchain"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/state"
	"BlockChain/src/utils"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
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
	log            *log.Logger
}

// NewBlockPool create a new block pool
func NewBlockPool(net *p2pnet.P2PNet, chain *blockchain.Chain, s *state.WorldState, logPath string) *BlockPool {
	// initialize logger
	l := utils.NewLogger("[BlockPool] ", logPath)

	if net == nil {
		l.Panic("unknown network")
	}

	pool := &BlockPool{
		pool:           make(map[string]*blockchain.Block),
		network:        net,
		chain:          chain,
		ws:             s,
		peerBestHeight: 0,
		newBlock:       make(chan *BlockMessage),
		syncMsg:        make(chan *BlockMessage),
		log:            l,
	}
	return pool
}

func (bp *BlockPool) Run() {
	bp.log.Println("Run Block Pool")
	// register receive callback func
	bp.network.RegisterCallback(p2pnet.BlockMsg, bp.OnReceive)

	// run block sync
	bp.log.Println("Begin Block synchronization")
	go bp.BlockSynchronization()

	// routine wait new block
	for {
		select {
		// receive new block
		case newBlockMsg := <-bp.newBlock:
			msg, _ := newBlockMsg.SplitMessage()
			if newBlock, ok := msg.(NewBlockMessage); ok {
				bp.log.Println("receive new block")
				var block blockchain.Block
				err := utils.Deserialize(newBlock.Block, &block)
				if err != nil {
					bp.log.Println("Deserialize Block fail")
				}
				if !(bp.chain.AddBlock(&block)) {
					bp.log.Println("Add Block to chain fail, put it in the pool")
					bp.AddBlock(&block)
				} else {
					bp.log.Println("Add Block to chain successfully")
					bp.log.Println("Reindex pool")
					bp.Reindex()
				}
			}
		}
	}
}

// OnReceive handle message receive from peer
func (bp *BlockPool) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.BlockMsg {
		return
	}
	var blockMsg BlockMessage
	bp.log.Println("Receive a new message")
	err := json.Unmarshal(msgBytes, &blockMsg)
	if err != nil {
		bp.log.Println("Unmarshal message fail")
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
		bp.log.Println("Marshal message fail")
	}
	p2pMsg := p2pnet.Message{
		Type: p2pnet.BlockMsg,
		Data: data,
	}
	bp.log.Println("Broadcast sync message to peer")
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
					bp.log.Println("Receive a SyncRequest message")
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
						bp.log.Println("Marshal message fail")
					}
					p2pMsg = p2pnet.Message{
						Type: p2pnet.BlockMsg,
						Data: data,
					}
					// send response to node
					bp.log.Println("Send request message to peer")
					bp.network.BroadcastToPeer(&p2pMsg, request.NodeID)
				}
			case SyncResponseMsg:
				// receive other peer response
				if response, ok := msg.(SyncResponseMessage); ok {
					bp.log.Println("Receive a SyncResponse message")
					// check best height and update
					if response.BestHeight > bp.peerBestHeight {
						bp.UpdatePeerBestHeight(response.BestHeight, response.FromID)
					}
				}
			case BlockRequestMsg:
				// handle block request
				if requestedBlock, ok := msg.(BlockRequestMessage); ok {
					bp.log.Println("Receive a BlockRequest message")
					blocksInRange := blockchain.FindBlocksInRange(bp.chain, requestedBlock.Min, requestedBlock.Max)
					for _, block := range blocksInRange {
						serializedData, err := utils.Serialize(block)
						if err != nil {
							bp.log.Println("Serialize block fail")
						}
						blockResMsg := BlockResponseMessage{
							FromID: bp.ws.SelfID,
							ToID:   requestedBlock.NodeID,
							Height: block.Header.Height,
							Hash:   block.Header.Hash,
							Block:  serializedData,
						}
						blockMsg = BlockMessage{
							Type: BlockResponseMsg,
							Data: blockResMsg,
						}
						data, err = json.Marshal(blockMsg)
						if err != nil {
							bp.log.Println("Marshal message fail")
						}
						p2pMsg = p2pnet.Message{
							Type: p2pnet.BlockMsg,
							Data: data,
						}
						bp.log.Println("Send a BlockResponse message to peer: ", requestedBlock.NodeID)
						bp.network.BroadcastToPeer(&p2pMsg, requestedBlock.NodeID)
					}
				}
			case BlockResponseMsg:
				// handle block response
				// 取出BlockResponseMessage中的区块信息
				if response, ok := msg.(BlockResponseMessage); ok {
					bp.log.Println("Receive a BlockResponse message")
					var block blockchain.Block
					err = utils.Deserialize(response.Block, &block)
					if err != nil {
						bp.log.Println("Deserialize block fail")
					}
					if !(bp.chain.AddBlock(&block)) {
						bp.log.Println("Add Block to chain fail, put it in the pool")
						bp.AddBlock(&block)
					} else {
						bp.log.Println("Add Block to chain successfully")
						bp.log.Println("Reindex pool")
						bp.Reindex()
					}
				}
			default:

			}
		case <-syncTimer.C:
			// 计时器判断同步请求是否完成
			// 请求完成，开始向拥有最长链的节点发送BlockRequestMessage消息请求区块
			if bp.peerBestHeight > bp.chain.BestHeight {
				bp.log.Println("Send Request message to peer: ", bp.bestPeerID)
				blockReqMsg := BlockRequestMessage{
					NodeID: bp.ws.SelfID,
					Min:    bp.chain.BestHeight + 1,
					Max:    bp.peerBestHeight,
				}
				blockMsg = BlockMessage{
					Type: BlockRequestMsg,
					Data: blockReqMsg,
				}
				data, err = json.Marshal(blockMsg)
				if err != nil {
					return
				}
				p2pMsg = p2pnet.Message{
					Type: p2pnet.BlockMsg,
					Data: data,
				}
				bp.network.BroadcastToPeer(&p2pMsg, bp.bestPeerID)
			}
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

// Reindex add orphan block to chain
func (bp *BlockPool) Reindex() {
	if bp.Count() != 0 {
		bp.mu.Lock()
		for _, block := range bp.pool {
			if bytes.Equal(block.Header.PrevHash, bp.chain.Tip) {
				bp.chain.AddBlock(block)
			}
		}
		bp.mu.Unlock()
	}
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
