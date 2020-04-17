package main

import (
	"log"
	"os"

	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/joho/godotenv"
)

type configurationValues struct {
	EncryptionPassphrase string
	KeyID                string
	ApplicationKey       string
	BucketID             string
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	configValues := getEnvironmentVariables()

	authorizationInfo := accountauthorization.GetAccountAuthorization(configValues.KeyID, configValues.ApplicationKey)

	goblaze.UploadDirectory(os.Args[1], configValues.EncryptionPassphrase, configValues.BucketID, authorizationInfo)

	downloadOptions := goblaze.DownloadOptions{
		DirectoryName:   "/home/jack/Documents/Backup-Test/",
		TargetDirectory: "/home/jack/Backup-Downloaded",
		WriteMode:       goblaze.OverwriteOldFiles,
	}

	goblaze.DownloadDirectory(downloadOptions, configValues.EncryptionPassphrase, authorizationInfo)
}

func getEnvironmentVariables() configurationValues {
	return configurationValues{
		EncryptionPassphrase: os.Getenv("ENCRYPTION_PASSPHRASE"),
		KeyID:                os.Getenv("KEY_ID"),
		ApplicationKey:       os.Getenv("APPLICATION_KEY"),
		BucketID:             os.Getenv("BUCKET_ID"),
	}
}
