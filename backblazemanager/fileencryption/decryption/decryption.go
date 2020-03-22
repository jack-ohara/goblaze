package decryption

import (
	"crypto/aes"
	"crypto/cipher"
	"log"

	"github.com/jack-ohara/backblazemanager/backblazemanager/fileencryption/passwordhasher"
)

func DecryptData(data []byte, passphrase string) []byte {
	key := passwordhasher.CreatePasswordHash(passphrase)

	block, err := aes.NewCipher(key)

	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		log.Fatal(err)
	}

	nonceSize := gcm.NonceSize()

	nonce, cipherText := data[:nonceSize], data[nonceSize:]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)

	if err != nil {
		log.Fatal(err)
	}

	return plainText
}
