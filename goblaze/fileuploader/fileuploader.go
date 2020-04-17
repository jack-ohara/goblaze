package fileuploader

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
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

type startLargeFileResponse struct {
	AccountID       string
	Action          string
	BucketID        string
	ContentLength   int
	ContentSha1     string
	ContentType     string
	FileID          string
	FileName        string
	UploadTimestamp time.Time
	StatusCode      int
}

type getUploadPartURLResponse struct {
	FileID             string
	UploadURL          string
	AuthorizationToken string
	StatusCode         int
}

type uploadPartResponse struct {
	FileID          string
	PartNumber      int
	ContentLength   int
	ContentSha1     string
	UploadTimestamp time.Time
	StatusCode      int
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
	LargeFile       bool
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

	uploadFileResponse := UploadFileResponse{StatusCode: response.StatusCode, LargeFile: false}

	if uploadFileResponse.StatusCode != 200 {
		log.Printf("Upload file failed with status code %d. Error: %s\n", uploadFileResponse.StatusCode, string(response.BodyContent))
	} else {
		json.Unmarshal(response.BodyContent, &uploadFileResponse)

		log.Println("Successfully uploaded ", filePath)
	}

	return uploadFileResponse
}

func uploadLargeFile(filePath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) UploadFileResponse {
	encryptedFileContents := encryption.EncryptFile(filePath, encryptionPassphrase)

	hash := sha1.New()

	hash.Write(encryptedFileContents)

	fileSha1 := hex.EncodeToString(hash.Sum(nil))

	startLargeFileResponse := startLargeFile(filePath, bucketID, fileSha1, authorizationInfo)

	if startLargeFileResponse.StatusCode != 200 {
		log.Printf("Start upload failed. See above for error message. Aborting upload attempt for %s\n", filePath)

		return UploadFileResponse{StatusCode: startLargeFileResponse.StatusCode}
	}

	var fileParts [][]byte

	numberOfParts := int64(math.Ceil(float64(len(encryptedFileContents)) / float64(authorizationInfo.RecommendedPartSize)))

	for i := int64(0); i < numberOfParts; i++ {
		var partContent []byte

		if i+1 == numberOfParts {
			partContent = encryptedFileContents[i*authorizationInfo.RecommendedPartSize:]
		} else {
			partContent = encryptedFileContents[i*authorizationInfo.RecommendedPartSize : (i+1)*authorizationInfo.RecommendedPartSize]
		}

		fileParts = append(fileParts, partContent)
	}

	var wg sync.WaitGroup
	var partSha1s = make([]string, numberOfParts)

	for index, part := range fileParts {
		wg.Add(1)

		go performPartUpload(startLargeFileResponse.FileID, authorizationInfo, index+1, part, &partSha1s, &wg)
	}

	wg.Wait()
	return finishLargeFile(startLargeFileResponse.FileID, authorizationInfo, partSha1s)
}

func startLargeFile(fileName, bucketID, fileSha1 string, authorizationInfo accountauthorization.AuthorizeAccountResponse) startLargeFileResponse {
	url := authorizationInfo.APIURL + "/b2api/v2/b2_start_large_file"

	headers := map[string]string{
		"Authorization": authorizationInfo.AuthorizationToken,
	}

	uploadFileName := fileName

	if strings.HasPrefix(fileName, "/") || strings.HasPrefix(fileName, "\\") {
		uploadFileName = fileName[1:]
	}

	type startLargeFileFileInfo struct {
		LargeFileSha1 string `json:"large_file_sha1"`
	}

	type startLargeFileBody struct {
		FileName    string                 `json:"fileName"`
		BucketID    string                 `json:"bucketId"`
		ContentType string                 `json:"contentType"`
		FileInfo    startLargeFileFileInfo `json:"fileInfo"`
	}

	requestBody := startLargeFileBody{
		FileName:    uploadFileName,
		BucketID:    bucketID,
		ContentType: "b2/x-auto",
		FileInfo:    startLargeFileFileInfo{LargeFileSha1: fileSha1},
	}

	body, _ := json.Marshal(requestBody)

	response := httprequestbuilder.ExecutePost(url, body, headers)

	startLargeFileResponse := startLargeFileResponse{StatusCode: response.StatusCode}

	if startLargeFileResponse.StatusCode != 200 {
		log.Printf("Start upload of file %s failed with error code %d. Error: %s\n", fileName, response.StatusCode, string(response.BodyContent))
	} else {
		json.Unmarshal(response.BodyContent, &startLargeFileResponse)
	}

	return startLargeFileResponse
}

func performPartUpload(fileID string, authorizationInfo accountauthorization.AuthorizeAccountResponse, partNumber int, partContents []byte, partSha1s *[]string, wg *sync.WaitGroup) {
	getUploadPartURLResponse := getUploadPartURL(fileID, authorizationInfo)

	uploadResponse := uploadPart(getUploadPartURLResponse.UploadURL, getUploadPartURLResponse.AuthorizationToken, partNumber, partContents)

	(*partSha1s)[partNumber-1] = uploadResponse.ContentSha1

	(*wg).Done()
}

func getUploadPartURL(fileID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) getUploadPartURLResponse {
	url := authorizationInfo.APIURL + "/b2api/v2/b2_get_upload_part_url"

	headers := map[string]string{"Authorization": authorizationInfo.AuthorizationToken}

	body, _ := json.Marshal(map[string]string{
		"fileId": fileID,
	})

	response := httprequestbuilder.ExecutePost(url, body, headers)

	getUploadPartURLResponse := getUploadPartURLResponse{StatusCode: response.StatusCode}

	if getUploadPartURLResponse.StatusCode != 200 {
		log.Printf("GetUploadPartUrl failed with status code %d. Error: %s\n", response.StatusCode, string(response.BodyContent))
	} else {
		json.Unmarshal(response.BodyContent, &getUploadPartURLResponse)
	}

	return getUploadPartURLResponse
}

func uploadPart(uploadPartURL, uploadPartURLAuthToken string, partNumber int, partContent []byte) uploadPartResponse {
	hash := sha1.New()

	hash.Write(partContent)

	partSha1 := hex.EncodeToString(hash.Sum(nil))

	headers := map[string]string{
		"Authorization":     uploadPartURLAuthToken,
		"X-Bz-Part-Number":  strconv.FormatInt(int64(partNumber), 10),
		"X-Bz-Content-Sha1": partSha1,
		"Content-Length":    strconv.FormatInt(int64(len(partContent)), 10),
	}

	response := httprequestbuilder.ExecutePost(uploadPartURL, partContent, headers)

	uploadPartResponse := uploadPartResponse{StatusCode: response.StatusCode}

	if uploadPartResponse.StatusCode != 200 {
		log.Printf("UploadPart failed with status code %d. Error: %s\n", uploadPartResponse.StatusCode, string(response.BodyContent))
	} else {
		json.Unmarshal(response.BodyContent, &uploadPartResponse)
	}

	return uploadPartResponse
}

func finishLargeFile(fileID string, authorizationInfo accountauthorization.AuthorizeAccountResponse, partSha1s []string) UploadFileResponse {
	url := authorizationInfo.APIURL + "/b2api/v2/b2_finish_large_file"

	headers := map[string]string{"Authorization": authorizationInfo.AuthorizationToken}

	type finishLargeFileBody struct {
		FileID        string   `json:"fileId"`
		PartSha1Array []string `json:"partSha1Array"`
	}

	bodyStruct := finishLargeFileBody{
		FileID:        fileID,
		PartSha1Array: partSha1s,
	}

	body, err := json.Marshal(bodyStruct)

	if err != nil {
		log.Println("Failed to marshal the finishLargeFile body: ", err)
	}

	response := httprequestbuilder.ExecutePost(url, body, headers)

	finishLargeFileResponse := UploadFileResponse{StatusCode: response.StatusCode, LargeFile: true}

	if response.StatusCode != 200 {
		log.Printf("Call to FinishLargeFile failed with code %d. Error: %s\n", response.StatusCode, string(response.BodyContent))
	} else {
		json.Unmarshal(response.BodyContent, &finishLargeFileResponse)
	}

	return finishLargeFileResponse
}
