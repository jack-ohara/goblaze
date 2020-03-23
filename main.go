package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/goblaze/filedownloader"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	encryptionPassphrase := os.Getenv("ENCRYPTION_PASSPHRASE")

	authorizationInfo := goblaze.GetAccountAuthorization(os.Getenv("KEY_ID"), os.Getenv("APPLICATION_ID"))

	//fileuploader.UploadFile("/home/jack/test.txt", encryptionPassphrase, authorizationInfo, os.Getenv("BUCKET_ID"))

	downloadResponse := filedownloader.DownloadFile("/home/jack/test.txt", os.Getenv("BUCKET_NAME"), authorizationInfo, encryptionPassphrase)

	fmt.Println(string(downloadResponse.FileContent))
}
