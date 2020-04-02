package goblaze

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
	"github.com/jack-ohara/goblaze/goblaze/uploadedfiles"
)

func UploadDirectories(directories []string, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedFiles := uploadedfiles.GetUploadedFiles()

	for _, directoryPath := range directories {
		for _, filePath := range getFilePaths(directoryPath) {
			if fileShouldBeUploaded(filePath, uploadedFiles) {
				uploadResponse := fileuploader.UploadFile(filePath, encryptionPassphrase, authorizationInfo, bucketID)

				if uploadResponse.StatusCode == 200 {
					uploadedFiles[filePath] = uploadedfiles.UploadedFileInfo{LastUploadedTime: time.Now(), FileID: uploadResponse.FileID}
				} else {
					log.Printf("The uploading of the file %s returned a status code of %d", filePath, uploadResponse.StatusCode)
				}
			}
		}
	}

	uploadedfiles.WriteUploadedFiles(uploadedFiles)
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

func fileShouldBeUploaded(filePath string, uploadedFiles uploadedfiles.UploadedFiles) bool {
	if uploadedFileInfo, fileHasBeenUploaded := uploadedFiles[filePath]; fileHasBeenUploaded {
		fileInfo, err := os.Stat(filePath)

		if err != nil {
			log.Fatal(err)
		}

		return uploadedFileInfo.LastUploadedTime.Before(fileInfo.ModTime().Local())
	}

	return true
}
