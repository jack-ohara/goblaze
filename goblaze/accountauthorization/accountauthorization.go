package accountauthorization

import (
	"encoding/base64"
	"encoding/json"

	"github.com/jack-ohara/goblaze/httprequestbuilder"
)

type AuthorizeAccountResponse struct {
	AccountID               string
	AuthorizationToken      string
	Allowed                 TokenCapabilities
	APIURL                  string
	DownloadURL             string
	RecommendedPartSize     int64
	AbsoluteMinimumPartSize int64
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

	resp := httprequestbuilder.ExecuteGet("https://api.backblazeb2.com/b2api/v2/b2_authorize_account", headers)

	json.Unmarshal(resp.BodyContent, &authorizeAccountRes)

	return authorizeAccountRes
}
