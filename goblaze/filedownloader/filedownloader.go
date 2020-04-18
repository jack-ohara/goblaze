package filedownloader

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/url"
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
	StatusCode  int
}

func DownloadFileByName(filename, decryptionPassphrase string, authorizationInfo accountauthorization.AuthorizeAccountResponse, largeFile bool) DownloadFileResponse {
	fileID := getFileID(filename)

	if fileID == "" {
		log.Fatalf("File %s has not previously been uploaded, so it can't be downloaded", filename)
	}

	return DownloadFileById(fileID, decryptionPassphrase, authorizationInfo, largeFile)
}

func DownloadFileById(fileID, decryptionPassphrase string, authorizationInfo accountauthorization.AuthorizeAccountResponse, largeFile bool) DownloadFileResponse {
	downloadURL := authorizationInfo.DownloadURL + "/b2api/v2/b2_download_file_by_id?fileId=" + fileID

	headers := map[string]string{
		"Authorization": authorizationInfo.AuthorizationToken,
	}

	response := httprequestbuilder.ExecuteGet(downloadURL, headers)

	downloadFileResponse := DownloadFileResponse{StatusCode: response.StatusCode}

	if downloadFileResponse.StatusCode != 200 {
		log.Printf("Download failed with response code %d. Error: %s\n", response.StatusCode, string(response.BodyContent))
	} else {
		encryptedContent := response.BodyContent

		hash := sha1.New()

		hash.Write(encryptedContent)

		fileSha1 := hex.EncodeToString(hash.Sum(nil))

		var sha1Header string

		if largeFile {
			sha1Header = "X-Bz-Info-Large_file_sha1"
		} else {
			sha1Header = "X-Bz-Content-Sha1"
		}

		if strings.Compare(response.Headers[sha1Header][0], fileSha1) != 0 {
			log.Println("Could not match the Sha1 for the file with ID: ", fileID)
		} else {
			fileContent := decryption.DecryptData(response.BodyContent, decryptionPassphrase)

			fileName := response.Headers["X-Bz-File-Name"][0]
			decoded, err := url.QueryUnescape(fileName)

			if err != nil {
				log.Printf("Error decoding the downloaded filename %s for file with ID %s\n", fileName, fileID)

				return downloadFileResponse
			}

			downloadFileResponse.FileID = response.Headers["X-Bz-File-Id"][0]
			downloadFileResponse.FileName = decoded
			downloadFileResponse.FileContent = fileContent
		}
	}

	return downloadFileResponse
}

func getFileID(filename string) string {
	uploadedFiles := uploadedfiles.GetUploadedFiles()

	if uploadedFileInfo, fileHasBeenUploaded := uploadedFiles[filename]; fileHasBeenUploaded {
		return uploadedFileInfo.FileID
	}

	return ""
}
