package uploadedfiles

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type UploadedFiles map[string]time.Time

func GetUploadedFiles() UploadedFiles {
	if _, err := os.Stat(getConfigDirectory()); os.IsNotExist(err) {
		os.MkdirAll(getConfigDirectory(), os.ModePerm)
	}

	if _, err := os.Stat(getConfigFilePath()); os.IsNotExist(err) {
		file, err := os.Create(getConfigFilePath())

		if err != nil {
			log.Fatal(err)
		}

		file.Close()
	}

	fileContents, err := ioutil.ReadFile(getConfigFilePath())

	if err != nil {
		log.Fatal(err)
	}

	uploadedFiles := make(UploadedFiles)
	json.Unmarshal(fileContents, &uploadedFiles)

	return uploadedFiles
}

func WriteUploadedFiles(uploadedFiles UploadedFiles) {
	jsonContent, err := json.Marshal(uploadedFiles)

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(getConfigFilePath(), jsonContent, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}

func getConfigDirectory() string {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	return path.Join(userHomeDir, ".goblaze")
}

func getConfigFilePath() string {
	return path.Join(getConfigDirectory(), "uploadedFiles.json")
}
