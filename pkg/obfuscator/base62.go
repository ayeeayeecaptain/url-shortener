package obfuscator

import (
	"errors"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const base = uint64(len(alphabet))

// Encode turns an integer ID into a Base62 token string
func Encode(id uint64) string {
	if id == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder
	for id > 0 {
		rem := id % base
		sb.WriteByte(alphabet[rem])
		id = id / base
	}

	// Reverse the string for standard encoding ordering
	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Decode translates a Base62 token string back into its original database sequential ID
func Decode(token string) (uint64, error) {
	var id uint64
	for i := 0; i < len(token); i++ {
		pos := strings.IndexByte(alphabet, token[i])
		if pos == -1 {
			return 0, errors.New("invalid character in base62 token")
		}
		id = id*base + uint64(pos)
	}
	return id, nil
}
