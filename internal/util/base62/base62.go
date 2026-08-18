package base62

import (
	"math/big"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base = int64(62)

// EncodeBytes encodes a byte slice into a Base62 string.
// It limits the output to `limit` characters by applying modulo 62^limit.
func EncodeBytes(b []byte, limit int) string {
	i := new(big.Int).SetBytes(b)
	return EncodeBigInt(i, limit)
}

// EncodeBigInt encodes a big.Int into a Base62 string up to `limit` characters.
func EncodeBigInt(i *big.Int, limit int) string {
	if i.Cmp(big.NewInt(0)) == 0 {
		return padLeft(string(alphabet[0]), limit)
	}

	// Calculate maximum allowed value: 62^limit
	baseInt := big.NewInt(base)
	maxVal := new(big.Int).Exp(baseInt, big.NewInt(int64(limit)), nil)

	// Apply modulo to ensure the number fits within the character limit
	val := new(big.Int).Mod(i, maxVal)

	if val.Cmp(big.NewInt(0)) == 0 {
		return padLeft(string(alphabet[0]), limit)
	}

	var result []byte
	zero := big.NewInt(0)
	mod := &big.Int{}

	for val.Cmp(zero) > 0 {
		val.DivMod(val, baseInt, mod)
		result = append(result, alphabet[mod.Int64()])
	}

	// Reverse the result because we extracted least significant digits first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return padLeft(string(result), limit)
}

func padLeft(str string, limit int) string {
	for len(str) < limit {
		str = string(alphabet[0]) + str
	}
	return str
}
