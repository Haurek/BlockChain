# BlockChian design

## 基础功能

### P2P网络

- 节点创建
- 节点发现
- 添加新交易到待处理池
- 提交区块到共识过程
- 获取特定交易或区块的信息
- 执行节点之间的同步操作(一致性)



### 存储

数据库？



### 共识机制

- PBFT--拜占庭问题
- PoW



### 节点和链数据结构

**block**

区块头：

- 时间戳
- 前一个块的hash
- 当前块hash
- nonce(验证PoW)
- Merkel tree root 
- height区块高度
- target_bits挖矿难度
- transaction_count包含交易个数

区块体：

- 数据(交易)



**chain**

- nodes





### 密码模块

ECDSA

非对称加密

数字签名



**地址**

由公钥生成，生成方式参考课程ppt

**钱包**

钱包只是一个密钥对，保存公钥和私钥



### UTXO

没有余额，每笔交易由至少一个输入和一个输出组成，一个新的交易的输入来自之前一笔交易的输出，某个地址所有未被引用的输出就是该地址的余额

UTXO输出：

- value
- 公钥hash

UTXO输入：

- id
- 第几个输出
- 签名
- 公钥

**coinbase**：只有输出没有输入，用于创世区块发行货币以及矿工挖矿奖励

reward是矿工挖矿的奖励，是coinbase类型，只有一个输出，其中包含矿工公钥hash

每笔交易由输入和输出组成，保存在Transaction数据结构，作为区块的区块体

余额计算只需要计算所有自己能解锁的UTXO输出value之和



### PoW

Hashcush：

SHA256(data+counter)值前n个是0，n是挖矿难度，counter从0开始递增



### client



### log

