package pool

import (
	"BlockChain/src/blockchain"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/utils"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
	"time"
)

const SyncPollingTime = 5

type BlockPool struct {
	full           int
	pool           map[string]*blockchain.Block
	chain          *blockchain.Chain
	network        *p2pnet.P2PNet
	peerBestHeight uint64
	bestPeerID     string
	newBlock       chan *BlockMessage
	syncMsg        chan *BlockMessage
	log            *log.Logger
	lock           sync.Mutex
}

// NewBlockPool create a new block pool
func NewBlockPool(f int, net *p2pnet.P2PNet, chain *blockchain.Chain, logPath string) *BlockPool {
	// initialize logger
	l := utils.NewLogger("[BlockPool] ", logPath)

	if net == nil {
		l.Panic("unknown network")
	}

	pool := &BlockPool{
		full:           f,
		pool:           make(map[string]*blockchain.Block),
		network:        net,
		chain:          chain,
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
}

// OnReceive handle message receive from peer
func (bp *BlockPool) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.BlockMsg {
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

	default:
		return
	}
}

func (bp *BlockPool) BlockSynchronization() {
	blockMsg, err := CreateBlockMessage(SyncRequestMsg, bp.network.ID, bp.chain.BestHeight)
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
	bp.BlockSynchronization()
	syncTimer := time.NewTimer(SyncPollingTime * time.Second)

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
					blockMsg, err := CreateBlockMessage(SyncResponseMsg, bp.network.ID, request.NodeID, bp.chain.BestHeight)

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
						blockMsg, err := CreateBlockMessage(BlockResponseMsg, bp.network.ID, requestedBlock.NodeID, block.Header.Height, block.Header.Hash, serializedData)
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
				blockMsg, err := CreateBlockMessage(BlockRequestMsg, bp.network.ID, bp.chain.BestHeight+1, bp.peerBestHeight)
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
				bp.BlockSynchronization()
				syncTimer.Reset(10 * time.Second)
			}
		}
	}
}

// AddBlock add block to pool
func (bp *BlockPool) AddBlock(block *blockchain.Block) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	id := hex.EncodeToString(block.Header.Hash)
	if _, exists := bp.pool[id]; !exists {
		// add block to pool
		bp.pool[id] = block
		// check pool status
		if len(bp.pool) >= bp.full {
			// add block to chain
			bp.Reindex()
		}
	}
}

// Reindex add orphan block to chain
func (bp *BlockPool) Reindex() {
	if bp.Count() != 0 {
		for {
			found := false
			for _, block := range bp.pool {
				if bytes.Equal(block.Header.PrevHash, bp.chain.Tip) {
					bp.chain.AddBlock(block)
					bp.RemoveBlock(block.Header.Hash)
					found = true
					break
				} else {
					continue
				}
			}
			if len(bp.pool) == 0 || !found {
				break
			}
		}
	}
}

// GetBlock get block from pool by hash
func (bp *BlockPool) GetBlock(hash []byte) *blockchain.Block {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	id := hex.EncodeToString(hash)
	if _, exists := bp.pool[id]; !exists {
		return bp.pool[id]
	}
	return nil
}

// RemoveBlock remove block from pool by hash
func (bp *BlockPool) RemoveBlock(hash []byte) {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	id := hex.EncodeToString(hash)
	if _, exists := bp.pool[id]; exists {
		delete(bp.pool, id)
	}
}

// HaveBlock check a block in pool
func (bp *BlockPool) HaveBlock(hash []byte) bool {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	id := hex.EncodeToString(hash)
	_, exists := bp.pool[id]
	return exists
}

// Count block num in pool
func (bp *BlockPool) Count() int {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	return len(bp.pool)
}

func (bp *BlockPool) UpdatePeerBestHeight(height uint64, id string) {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	bp.peerBestHeight = height
	bp.bestPeerID = id
}
