package fileencryption_test

import (
	"bytes"
	"testing"

	"github.com/jack-ohara/goblaze/goblaze/fileencryption/decryption"
	"github.com/jack-ohara/goblaze/goblaze/fileencryption/encryption"
)

func TestEncryptionAndDecryption(t *testing.T) {
	data := []byte("this is some test data that will be encrypted! And then decrypted!")
	passphrase := "Some_Really$ecure>P@ssphr@Se!"

	encryptedData := encryption.EncryptData(data, passphrase)

	if bytes.Compare(encryptedData, data) == 0 {
		t.Error("The encrypted data is the same as the plain text data that was passed in")
	}

	decryptedData := decryption.DecryptData(encryptedData, passphrase)

	if bytes.Compare(decryptedData, data) != 0 {
		t.Errorf("Decrypted data = %s; expected %s", decryptedData, data)
	}
}
