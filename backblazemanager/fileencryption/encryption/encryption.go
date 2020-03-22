package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"

	"github.com/jack-ohara/backblazemanager/backblazemanager/fileencryption/passwordhasher"
)

func EncryptFile(filePath, passphrase string) []byte {
	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		log.Fatal(err)
	}

	return EncryptData(data, passphrase)
}

func EncryptData(data []byte, passphrase string) []byte {
	passwordHash := passwordhasher.CreatePasswordHash(passphrase)
	block, _ := aes.NewCipher(passwordHash)

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		log.Fatal(err)
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	return gcm.Seal(nonce, nonce, data, nil)
}
