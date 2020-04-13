package main

import (
	"github.com/jack-ohara/goblaze/httprequestbuilder"
	"log"
	"os"

	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	encryptionPassphrase := os.Getenv("ENCRYPTION_PASSPHRASE")

	authorizationInfo := accountauthorization.GetAccountAuthorization(os.Getenv("KEY_ID"), os.Getenv("APPLICATION_ID"))

	goblaze.UploadDirectories(os.Args[1:], encryptionPassphrase, os.Getenv("BUCKET_ID"), authorizationInfo)

	httprequestbuilder.ExecutePost(authorizationInfo.APIURL + "/b2api/v2/b2_list_file_names", []byte("{\"bucketId\": \""+os.Getenv("BUCKET_ID")+"\"}"), map[string]string{"Authorization": authorizationInfo.AuthorizationToken})

	// fileuploader.UploadFile("/home/jack/test.txt", encryptionPassphrase, authorizationInfo, os.Getenv("BUCKET_ID"))

	// downloadResponse := filedownloader.DownloadFile("/home/jack/Documents/Backup-Test/file1.txt", authorizationInfo, encryptionPassphrase)

	// fmt.Println(string(downloadResponse.FileContent))
}
