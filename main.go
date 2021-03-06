package main

import (
	"flag"
	"log"
	"os"

	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/configuration"
)

func main() {
	configCommand := flag.NewFlagSet("config", flag.ExitOnError)

	uploadCommand := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadDirectory := uploadCommand.String("dir", "", "Identifies the directory to upload")

	downloadCommand := flag.NewFlagSet("download", flag.ExitOnError)
	downloadDirectory := downloadCommand.String("dir", "", "Identifies the directory to download from backblaze")
	downloadDestination := downloadCommand.String("dest", ".", "Identifies the location on disk that the downloaded files will be written to")
	downloadWriteMode := downloadCommand.Int("write-mode", 1, "Value of 0: Does not overwrite the file if it already exists\nValue of 1: Overwrites existing files if the downloaded file is more recent\nValue of 2: Overwrites any existing files")

	switch os.Args[1] {
	case "config":
		configCommand.Parse(os.Args[2:])

		if len(configCommand.Args()) > 0 {
			log.Fatalln("Unexpected arguments to goblaze config: ", uploadCommand.Args())
		}

		configuration.SetupConfigFile()
	case "upload":
		uploadCommand.Parse(os.Args[2:])

		if len(uploadCommand.Args()) > 0 {
			log.Fatalln("Unexpected arguments to goblaze upload: ", uploadCommand.Args())
		}

		configValues := configuration.GetConfiguration()

		fileInfo, err := os.Stat(*uploadDirectory)

		if os.IsNotExist(err) {
			log.Fatalln("Directory does not exist: ", *uploadDirectory)
		}

		if !fileInfo.IsDir() {
			log.Fatalln("Expected 'dir' argument to point to a directory but it is a file: ", *uploadDirectory)
		}

		authorizationInfo := accountauthorization.GetAccountAuthorization(configValues.KeyID, configValues.ApplicationKey)

		goblaze.UploadDirectory(*uploadDirectory, configValues.EncryptionPassphrase, configValues.BucketID, authorizationInfo)
	case "download":
		downloadCommand.Parse(os.Args[2:])

		if len(downloadCommand.Args()) > 0 {
			log.Fatalln("Unexpected arguments to goblaze download: ", downloadCommand.Args())
		}

		configValues := configuration.GetConfiguration()

		fileInfo, err := os.Stat(*downloadDestination)

		if os.IsNotExist(err) {
			log.Fatalln("Destination directory does not exist: ", *downloadDestination)
		}

		if err != nil {
			log.Fatalln(err)
		}

		if !fileInfo.IsDir() {
			log.Fatalln("Expected 'dest' argument to point to a directory but it is a file: ", *downloadDestination)
		}

		if *downloadWriteMode != 0 && *downloadWriteMode != 1 && *downloadWriteMode != 2 {
			log.Fatalln("Invalid value for write-mode: ", *downloadWriteMode)
		}

		authorizationInfo := accountauthorization.GetAccountAuthorization(configValues.KeyID, configValues.ApplicationKey)

		downloadOptions := goblaze.DownloadOptions{
			DirectoryName:   *downloadDirectory,
			TargetDirectory: *downloadDestination,
			WriteMode:       goblaze.FileWriteMode(*downloadWriteMode),
		}

		goblaze.DownloadDirectory(downloadOptions, configValues.EncryptionPassphrase, authorizationInfo)
	default:
		log.Fatalln("Expected a subcommand of 'upload' or 'download'")
	}
}
