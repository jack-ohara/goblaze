package filedownloader

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"strings"

	"github.com/jack-ohara/goblaze/fileencryption/decryption"
	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/uploadedfiles"
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
func DownloadFile(filename string, authorizeAccountResponse accountauthorization.AuthorizeAccountResponse, decryptionPassphrase string) DownloadFileResponse {
	fileID := getFileID(filename)

	if fileID == "" {
		log.Fatalf("File %s has not previously been uploaded, so it can't be downloaded", filename)
	}

	url := authorizeAccountResponse.DownloadURL + "/b2api/v2/b2_download_file_by_id?fileId=" + fileID

	headers := map[string]string{
		"Authorization": authorizeAccountResponse.AuthorizationToken,
	}

	response := httprequestbuilder.ExecuteGet(url, headers)

	if response.StatusCode != 200 {
		log.Fatalf("Download failed with response code %d. Error: %s", response.StatusCode, string(response.BodyContent))
	}

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

func getFileID(filename string) string {
	uploadedFiles := uploadedfiles.GetUploadedFiles()

	if uploadedFileInfo, fileHasBeenUploaded := uploadedFiles[filename]; fileHasBeenUploaded {
		return uploadedFileInfo.FileID
	}

	return ""
}
