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
	go bp.BlockSyncRoutine()

	// routine wait new block
	for {
		select {
		// receive new block
		case newBlockMsg := <-bp.newBlock:
			msg, _ := newBlockMsg.SplitMessage()
			if newBlock, ok := msg.(NewBlockMessage); ok {
				bp.log.Println("receive new block")
				var block blockchain.Block
				err := json.Unmarshal(newBlock.Block, &block)
				if err != nil {
					bp.log.Println("Deserialize Block fail")
				}
				bp.log.Printf("block height: %d, block hash: %s", block.Header.Height, block.Header.Hash)
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
		bp.log.Println("Unknown message type")
		return
	}
	var blockMsg BlockMessage
	err := json.Unmarshal(msgBytes, &blockMsg)
	if err != nil {
		bp.log.Println("Unmarshal message fail")
		return
	}
	// handle message receive
	switch blockMsg.Type {
	case SyncRequestMsg:
		fallthrough
	case SyncResponseMsg:
		fallthrough
	case BlockRequestMsg:
		fallthrough
	case BlockResponseMsg:
		bp.syncMsg <- &blockMsg
		bp.log.Println("Receive a sync message")

	case NewBlockBroadcastMsg:
		bp.newBlock <- &blockMsg
		bp.log.Println("Receive a new block message")

	default:
		return
	}
}

func (bp *BlockPool) BlockSynchronization() {
	blockMsg, err := CreateBlockMessage(SyncRequestMsg, bp.ws.SelfID, bp.chain.BestHeight)
	if err != nil {
		bp.log.Println("Create message fail")
		return
	}
	data, err := json.Marshal(blockMsg)
	if err != nil {
		bp.log.Println("Marshal message fail")
		return
	}
	p2pMsg := p2pnet.Message{
		Type: p2pnet.BlockMsg,
		Data: data,
	}
	bp.log.Printf("Broadcast sync message, chain height: %d", bp.chain.BestHeight)
	bp.network.Broadcast(&p2pMsg)
}

func (bp *BlockPool) BlockSyncRoutine() {
	go bp.BlockSynchronization()
	syncTimer := time.NewTimer(5 * time.Second)

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
					blockMsg, err := CreateBlockMessage(SyncResponseMsg, bp.ws.SelfID, request.NodeID, bp.chain.BestHeight)

					data, err := json.Marshal(blockMsg)
					if err != nil {
						bp.log.Println("Marshal message fail")
						break
					}
					p2pMsg := p2pnet.Message{
						Type: p2pnet.BlockMsg,
						Data: data,
					}
					// send response to node
					bp.log.Println("Send response message to peer")
					bp.network.BroadcastToPeer(&p2pMsg, request.NodeID)
				}
			case SyncResponseMsg:
				// receive other peer response
				if response, ok := msg.(SyncResponseMessage); ok {
					bp.log.Println("Receive a SyncResponse message")
					// check best height and update
					if response.BestHeight > bp.peerBestHeight {
						bp.log.Println("Update bestheight to: ", response.BestHeight)
						bp.UpdatePeerBestHeight(response.BestHeight, response.FromID)
					}
				}
			case BlockRequestMsg:
				// handle block request
				if requestedBlock, ok := msg.(BlockRequestMessage); ok {
					bp.log.Println("Receive a BlockRequest message")
					blocksInRange := bp.chain.FindBlocksInRange(requestedBlock.Min, requestedBlock.Max)
					for _, block := range blocksInRange {
						serializedData, err := json.Marshal(block)
						if err != nil {
							bp.log.Println("Marshal block fail")
							continue
						}

						// send block response message
						blockMsg, err := CreateBlockMessage(BlockResponseMsg, bp.ws.SelfID, requestedBlock.NodeID, block.Header.Height, block.Header.Hash, serializedData)
						data, err := json.Marshal(blockMsg)
						if err != nil {
							bp.log.Println("Marshal message fail")
							continue
						}
						p2pMsg := p2pnet.Message{
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
					err := json.Unmarshal(response.Block, &block)
					if err != nil {
						bp.log.Println("Deserialize block fail")
						break
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
			// sync timeout
			if bp.peerBestHeight > bp.chain.BestHeight {
				bp.log.Println("Send block request message to peer: ", bp.bestPeerID)
				blockMsg, err := CreateBlockMessage(BlockRequestMsg, bp.ws.SelfID, bp.chain.BestHeight+1, bp.peerBestHeight)
				data, err := json.Marshal(blockMsg)
				if err != nil {
					bp.log.Println("Marshal block fail")
					syncTimer.Reset(10 * time.Second)
					break
				}
				p2pMsg := p2pnet.Message{
					Type: p2pnet.BlockMsg,
					Data: data,
				}
				bp.network.BroadcastToPeer(&p2pMsg, bp.bestPeerID)
			} else {
				go bp.BlockSynchronization()
				syncTimer.Reset(10 * time.Second)
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
