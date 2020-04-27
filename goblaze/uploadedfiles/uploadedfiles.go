package uploadedfiles

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jack-ohara/goblaze/goblaze/configuration"
)

type UploadedFileInfo struct {
	LastUploadedTime time.Time
	FileID           string
	LargeFile        bool
}

type UploadedFiles map[string]UploadedFileInfo

func GetUploadedFiles() UploadedFiles {
	if _, err := os.Stat(configuration.GetUploadedFilesPath()); os.IsNotExist(err) {
		file, err := os.Create(configuration.GetUploadedFilesPath())

		if err != nil {
			log.Fatal(err)
		}

		file.Close()
	}

	fileContents, err := ioutil.ReadFile(configuration.GetUploadedFilesPath())

	if err != nil {
		log.Fatal(err)
	}

	uploadedFiles := make(UploadedFiles)
	json.Unmarshal(fileContents, &uploadedFiles)

	return uploadedFiles
}

func WriteUploadedFiles(uploadedFiles UploadedFiles) {
	jsonContent, err := json.MarshalIndent(uploadedFiles, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(configuration.GetUploadedFilesPath(), jsonContent, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}
