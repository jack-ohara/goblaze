package goblaze

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
)

type uploadedFiles map[string]time.Time

func UploadDirectories(directories []string, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedFiles := getUploadedFiles()

	for _, directoryPath := range directories {
		for _, filePath := range getFilePaths(directoryPath) {
			if fileShouldBeUploaded(filePath, uploadedFiles) {
				uploadResponse := fileuploader.UploadFile(filePath, encryptionPassphrase, authorizationInfo, bucketID)

				if uploadResponse.StatusCode == 200 {
					uploadedFiles[filePath] = time.Now()
				} else {
					log.Printf("The uploading of the file %s returned a status code of %d", filePath, uploadResponse.StatusCode)
				}
			}
		}
	}

	writeUploadedFiles(uploadedFiles)
}

func getFilePaths(directoryPath string) []string {
	var files []string

	allFiles, err := ioutil.ReadDir(directoryPath)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range allFiles {
		if file.IsDir() {
			files = append(files, getFilePaths(path.Join(directoryPath, file.Name()))...)
		} else {
			files = append(files, path.Join(directoryPath, file.Name()))
		}
	}

	return files
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

func getUploadedFiles() uploadedFiles {
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

	uploadedFiles := make(uploadedFiles)
	json.Unmarshal(fileContents, &uploadedFiles)

	return uploadedFiles
}

func fileShouldBeUploaded(filePath string, uploadedFiles uploadedFiles) bool {
	if lastUploadedTime, fileHasBeenUploaded := uploadedFiles[filePath]; fileHasBeenUploaded {
		fileInfo, err := os.Stat(filePath)

		if err != nil {
			log.Fatal(err)
		}

		return lastUploadedTime.Before(fileInfo.ModTime().Local())
	}

	return true
}

func writeUploadedFiles(uploadedFiles uploadedFiles) {
	jsonContent, err := json.Marshal(uploadedFiles)

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(getConfigFilePath(), jsonContent, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}
