package utils

//import (
//	"bytes"
//	"math"
//	"math/big"
//)
//
//func Mining(header *BlockHeader) (int64, []byte) {
//	var hashInt big.Int
//	headerBytes := SerializeHeader(header)
//	target := big.NewInt(1)
//	target.Lsh(target, uint(256-header.Bits))
//
//	var nonce int64
//	var hash []byte
//	nonce = 0
//	for nonce < math.MaxInt64 {
//		data := append(headerBytes, byte(nonce))
//		// hash twice
//		hash = Sha256Hash(data)
//		hash = Sha256Hash(hash)
//		hashInt.SetBytes(hash)
//		if hashInt.Cmp(target) == -1 {
//			break
//		} else {
//			nonce++
//		}
//	}
//	if nonce != math.MaxInt64 {
//		return nonce, hash[:]
//	} else {
//		panic("fail to mine block")
//	}
//}
//
//// Proof check hash
//func Proof(header *BlockHeader) bool {
//	headerBytes := SerializeHeader(header)
//	nonce := byte(header.Nonce)
//	headerBytes = append(headerBytes, nonce)
//	hash := Sha256Hash(headerBytes)
//	hash = Sha256Hash(hash[:])
//	return bytes.Equal(hash[:], header.Hash)
//}
//
//func SerializeHeader(header *BlockHeader) []byte {
//	data := bytes.Join(
//		[][]byte{
//			header.PrevHash,
//			header.MerkleRoot,
//			Int2Bytes(int64(header.Timestamp)),
//			Int2Bytes(int64(header.Bits)),
//		},
//		[]byte{},
//	)
//	return data
//}
