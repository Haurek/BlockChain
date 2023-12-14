# BlockChian design

## 共识层

共识层处理节点向主节点发送打包区块请求到区块加入区块链的整个过程，整个过程包括：

1. 客户端（区块链中的一个节点，不是主节点）通过**命令行**向主节点发起一个打包区块请求，发送RequestMessage到主节点，其中包括要打包交易的数据
2. 主节点接收消息后对交易进行验证，成功后广播给Pre-prepareMessage，然后进入Prepare状态等待其他节点广播的Prepare消息；客户端节点不参与共识过程，进入Reply阶段等待其他节点的回复

## 区块链层



## 网络层

libp2p框架下实现的P2P网络，网络间传递的消息类型`message.go/Message struct`，封装消息类型和具体的高层消息

消息类型：

```
const (
	BlockMsg       MessageType = iota // 区块链同步和区块广播消息
	TransactionMsg // 交易广播消息
	ConsensusMsg  // 共识层消息
)
```

消息处理过程：`handler.go/handleStream()`函数中通过线程`recvData`和`sendData`处理两个节点间的消息接收和发送，接收的消息根据其消息类型调用对应的回调函数处理
