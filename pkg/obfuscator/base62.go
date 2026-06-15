package obfuscator

import (
	"errors"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const base = uint64(len(alphabet))

// Secret keys for our Feistel Cipher round functions (interviewer-worthy encryption concept)
const feistelKey1 = 0xbf58476d1ce4e5b9
const feistelKey2 = 0x94d049bb133111eb

// scramble uses a 2-round Feistel Cipher to map a 64-bit integer to a unique, pseudo-random 64-bit integer.
// This eliminates predictable sequential short URLs entirely.
func scramble(val uint64) uint64 {
	left := uint32(val >> 32)
	right := uint32(val & 0xFFFFFFFF)

	// Round 1
	left ^= uint32((uint64(right) ^ feistelKey1) % 0xFFFFFFFF)
	// Round 2
	right ^= uint32((uint64(left) ^ feistelKey2) % 0xFFFFFFFF)

	return (uint64(left) << 32) | uint64(right)
}

// descramble reverses the Feistel Cipher operations in the exact opposite order to reveal the database sequential ID.
func descramble(val uint64) uint64 {
	left := uint32(val >> 32)
	right := uint32(val & 0xFFFFFFFF)

	// Reverse Round 2
	right ^= uint32((uint64(left) ^ feistelKey2) % 0xFFFFFFFF)
	// Reverse Round 1
	left ^= uint32((uint64(right) ^ feistelKey1) % 0xFFFFFFFF)

	return (uint64(left) << 32) | uint64(right)
}

// Encode scrambles the sequential database ID and transforms it into an unguessable Base62 token string.
func Encode(id uint64) string {
	if id == 0 {
		return string(alphabet[0])
	}

	// Scramble the ID before turning it into characters
	obfuscatedID := scramble(id)

	var sb strings.Builder
	for obfuscatedID > 0 {
		rem := obfuscatedID % base
		sb.WriteByte(alphabet[rem])
		obfuscatedID = obfuscatedID / base
	}

	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Decode translates the token string back into an integer, and descrambles it back to the original sequential database ID.
func Decode(token string) (uint64, error) {
	var obfuscatedID uint64
	for i := 0; i < len(token); i++ {
		pos := strings.IndexByte(alphabet, token[i])
		if pos == -1 {
			return 0, errors.New("invalid character in base62 token")
		}
		obfuscatedID = obfuscatedID*base + uint64(pos)
	}

	// Reverse the scramble to find the true autoincrement ID
	return descramble(obfuscatedID), nil
}
