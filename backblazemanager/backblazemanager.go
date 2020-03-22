package backblazemanager

import (
	"encoding/base64"
	"encoding/json"

	"github.com/jack-ohara/backblazemanager/httprequestbuilder"
)

type AuthorizeAccountResponse struct {
	AccountID               string
	AuthorizationToken      string
	Allowed                 TokenCapabilities
	APIURL                  string
	DownloadURL             string
	RecommendedPartSize     int
	AbsoluteMinimumPartSize int
}

type TokenCapabilities struct {
	BucketID     string
	BucketName   string
	Capabilities []string
	NamePrefix   string
}

func GetAccountAuthorization(keyID, applicationKey string) AuthorizeAccountResponse {
	appKeyHeader := base64.StdEncoding.EncodeToString([]byte(keyID + ":" + applicationKey))

	headers := map[string]string{
		"Authorization": "Basic " + appKeyHeader,
	}

	authorizeAccountRes := AuthorizeAccountResponse{}

	respBody := httprequestbuilder.ExecuteGet("https://api.backblazeb2.com/b2api/v2/b2_authorize_account", headers)

	json.Unmarshal(respBody, &authorizeAccountRes)

	return authorizeAccountRes
}
