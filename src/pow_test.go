package BlockChain

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"testing"
)

func TestMining(t *testing.T) {
	var hashInt big.Int
	target := big.NewInt(1)
	target.Lsh(target, uint(256-8))
	str := "12345678901234567890123456789099"
	headerBytes := []byte(str)
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
		fmt.Println(nonce)
		fmt.Println(string(hash[:]))
	}
}
