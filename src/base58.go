package BlockChain

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode encodes input to Base58 byte array
func Base58Encode(input []byte) []byte {
	var enc []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		enc = append(enc, b58Alphabet[mod.Int64()])
	}

	// bitcoin
	if input[0] == 0x00 {
		enc = append(enc, b58Alphabet[0])
	}

	for i, j := 0, len(enc)-1; i < j; i, j = i+1, j-1 {
		enc[i], enc[j] = enc[j], enc[i]
	}

	return enc
}

// Base58Decode decodes encoded data to byte array
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}
