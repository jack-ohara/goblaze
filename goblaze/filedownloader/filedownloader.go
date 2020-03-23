package filedownloader

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"strings"

	"github.com/jack-ohara/goblaze/fileencryption/decryption"
	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/httprequestbuilder"
)

type DownloadFileResponse struct {
	FileID      string
	FileName    string
	FileContent []byte
}

/*
	DownloadFile is a function for downloading and decrypting
	the specified file from	the specified bucket
*/
func DownloadFile(filename, bucketName string,
	authorizeAccountResponse goblaze.AuthorizeAccountResponse,
	decryptionPassphrase string) DownloadFileResponse {
	if strings.HasPrefix(filename, "/") {
		filename = filename[1:]
	}

	url := authorizeAccountResponse.APIURL + "/file/" + bucketName + "/" + filename

	headers := map[string]string{
		"Authorization": authorizeAccountResponse.AuthorizationToken,
	}

	response := httprequestbuilder.ExecuteGet(url, headers)

	encryptedContent := response.BodyContent

	hash := sha1.New()

	hash.Write(encryptedContent)

	fileSha1 := hex.EncodeToString(hash.Sum(nil))

	if strings.Compare(response.Headers["X-Bz-Content-Sha1"][0], fileSha1) != 0 {
		log.Fatalf("Could not match the Sha1 for the file %s", filename)
	}

	fileContent := decryption.DecryptData(response.BodyContent, decryptionPassphrase)

	return DownloadFileResponse{
		FileID:      response.Headers["X-Bz-File-Id"][0],
		FileName:    response.Headers["X-Bz-File-Name"][0],
		FileContent: fileContent,
	}
}
