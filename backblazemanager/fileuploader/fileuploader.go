package fileuploader

import (
	"encoding/json"

	"github.com/jack-ohara/backblazemanager/backblazemanager"
	"github.com/jack-ohara/backblazemanager/httprequestbuilder"
)

type GetUploadURLResponse struct {
	BucketID           string
	UploadURL          string
	AuthorizationToken string
}

// type UploadFileResponse struct {
// }

func UploadFile(filepath string, authorizationInfo backblazemanager.AuthorizeAccountResponse, bucketID string) GetUploadURLResponse {
	getUploadResponse := getUploadURL(authorizationInfo, bucketID)

	return getUploadResponse
	//return performUpload(filepath, getUploadResponse)
}

func getUploadURL(authInfo backblazemanager.AuthorizeAccountResponse, bucketID string) GetUploadURLResponse {
	url := authInfo.APIURL + "/b2api/v2/b2_get_upload_url"

	body, _ := json.Marshal(map[string]string{
		"bucketId": bucketID,
	})

	headers := map[string]string{
		"Authorization": authInfo.AuthorizationToken,
	}

	responseBody := httprequestbuilder.ExecutePost(url, body, headers)

	getUploadURLResponse := GetUploadURLResponse{}

	json.Unmarshal(responseBody, &getUploadURLResponse)

	return getUploadURLResponse
}

// func performUpload(filepath string, getUploadURLResponse GetUploadURLResponse) UploadFileResponse {
// 	return UploadFileResponse{}
// }
