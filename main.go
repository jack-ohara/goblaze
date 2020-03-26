package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/filedownloader"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	encryptionPassphrase := os.Getenv("ENCRYPTION_PASSPHRASE")

	authorizationInfo := accountauthorization.GetAccountAuthorization(os.Getenv("KEY_ID"), os.Getenv("APPLICATION_ID"))

	//fileuploader.UploadFile("/home/jack/test.txt", encryptionPassphrase, authorizationInfo, os.Getenv("BUCKET_ID"))

	bucketName := authorizationInfo.Allowed.BucketName

	if len(bucketName) == 0 {
		bucketName = os.Getenv("BUCKET_NAME")

		if len(bucketName) == 0 {
			log.Fatal("If you are not using a restricted access key, you must provide the BUCKET_NAME")
		}
	}

	downloadResponse := filedownloader.DownloadFile("home/jack/test.txt", authorizationInfo.Allowed.BucketName, authorizationInfo, encryptionPassphrase)

	fmt.Println(string(downloadResponse.FileContent))
}
