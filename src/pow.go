package BlockChain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
)

func Mining(header *BlockHeader) (int64, []byte) {
	var hashInt big.Int
	headerBytes := SerializeHeader(header)
	target := big.NewInt(1)
	target.Lsh(target, uint(256-header.Bits))

	var nonce int64
	var hash [32]byte
	nonce = 0
	for nonce < math.MaxInt64 {
		data := append(headerBytes, byte(nonce))
		// hash twice
		hash = sha256.Sum256(data)
		hash = sha256.Sum256(hash[:])
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(target) == -1 {
			break
		} else {
			nonce++
		}
	}
	if nonce != math.MaxInt64 {
		return nonce, hash[:]
	} else {
		panic("fail to mine block")
	}
}

func SerializeHeader(header *BlockHeader) []byte {
	data := bytes.Join(
		[][]byte{
			header.PrevHash,
			header.MerkleRoot,
			int2Bytes(int64(header.Timestamp)),
			int2Bytes(int64(header.Bits)),
		},
		[]byte{},
	)
	return data
}

func int2Bytes(value int64) []byte {
	buffer := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buffer, value)
	Bytes := make([]byte, n)
	copy(Bytes, buffer[:n])
	return Bytes
}
