package fileuploader

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jack-ohara/goblaze/fileencryption/encryption"
	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/httprequestbuilder"
)

type getUploadURLResponse struct {
	BucketID           string
	UploadURL          string
	AuthorizationToken string
	StatusCode         int
}

type UploadFileResponse struct {
	AccountID       string
	Action          string
	BucketID        string
	ContentLength   int
	ContentSha1     string
	FileID          string
	FileInfo        string
	FileName        string
	UploadTimestamp time.Time
	StatusCode      int
}

func UploadFile(filePath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) UploadFileResponse {
	fileInfo, err := os.Stat(filePath)

	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.Size() >= authorizationInfo.RecommendedPartSize {
		return uploadLargeFile(filePath, encryptionPassphrase, bucketID, authorizationInfo)
	}

	return uploadFile(filePath, encryptionPassphrase, bucketID, authorizationInfo)
}

func uploadFile(filePath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) UploadFileResponse {
	getUploadResponse := getUploadURL(authorizationInfo, bucketID)

	return performUpload(filePath, encryptionPassphrase, getUploadResponse)
}

func uploadLargeFile(filePath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) UploadFileResponse {
	return UploadFileResponse{}
}

func getUploadURL(authInfo accountauthorization.AuthorizeAccountResponse, bucketID string) getUploadURLResponse {
	url := authInfo.APIURL + "/b2api/v2/b2_get_upload_url"

	body, _ := json.Marshal(map[string]string{
		"bucketId": bucketID,
	})

	headers := map[string]string{
		"Authorization": authInfo.AuthorizationToken,
	}

	response := httprequestbuilder.ExecutePost(url, body, headers)

	getUploadURLResponse := getUploadURLResponse{StatusCode: response.StatusCode}

	json.Unmarshal(response.BodyContent, &getUploadURLResponse)

	return getUploadURLResponse
}

func performUpload(filePath, encryptionPassphrase string, getUploadURLResponse getUploadURLResponse) UploadFileResponse {
	encryptedFileContents := encryption.EncryptFile(filePath, encryptionPassphrase)

	hash := sha1.New()

	hash.Write(encryptedFileContents)

	fileSha1 := hex.EncodeToString(hash.Sum(nil))

	uploadFileName := filePath

	if strings.HasPrefix(filePath, "/") || strings.HasPrefix(filePath, "\\") {
		uploadFileName = filePath[1:]
	}

	headers := map[string]string{
		"Authorization":     getUploadURLResponse.AuthorizationToken,
		"X-Bz-File-Name":    uploadFileName,
		"Content-Type":      "b2/x-auto",
		"X-Bz-Content-Sha1": fileSha1,
	}

	response := httprequestbuilder.ExecutePost(getUploadURLResponse.UploadURL, encryptedFileContents, headers)

	uploadFileResponse := UploadFileResponse{StatusCode: response.StatusCode}

	if uploadFileResponse.StatusCode != 200 {
		log.Printf("Upload file failed with status code %d. Error: %s", uploadFileResponse.StatusCode, string(response.BodyContent))
	} else {
		json.Unmarshal(response.BodyContent, &uploadFileResponse)
	}

	return uploadFileResponse
}
