package passwordhasher

import (
	"crypto/sha256"
)

func CreatePasswordHash(key string) []byte {
	hash := sha256.Sum256([]byte(key))

	return hash[:]
}
