package BlockChain

import (
	"crypto/sha256"
	"encoding/binary"
	"goland/x/crypto/ripemd160"
)

func Int2Bytes(value int64) []byte {
	buffer := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buffer, value)
	Bytes := make([]byte, n)
	copy(Bytes, buffer[:n])
	return Bytes
}

// CalculateChecksum calculate checksum for address generate
func CalculateChecksum(payload []byte) []byte {
	firstHash := Sha256Hash(payload)
	secondHash := Sha256Hash(firstHash[:])
	// the first 4 byte are checksum
	return secondHash[:4]
}

func Sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func Ripemd160Hash(data []byte) []byte {
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(data)
	hash := ripemd160Hasher.Sum(nil)
	return hash[:]
}
