package goblaze

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
	"github.com/jack-ohara/goblaze/goblaze/uploadedfiles"
)


func UploadDirectories(directories []string, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedFiles := uploadedfiles.GetUploadedFiles()
	lock := sync.RWMutex{}

	var wg sync.WaitGroup

	for _, directoryPath := range directories {
		for _, filePath := range getFilePaths(directoryPath, uploadedFiles) {
			wg.Add(1)

			go uploadFile(filePath, encryptionPassphrase, bucketID, authorizationInfo, &lock, &uploadedFiles, &wg)
		}
	}

	wg.Wait()
	uploadedfiles.WriteUploadedFiles(uploadedFiles)
}

func getFilePaths(directoryPath string, uploadedFiles uploadedfiles.UploadedFiles) []string {
	var files []string

	allFiles, err := ioutil.ReadDir(directoryPath)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range allFiles {
		if file.IsDir() {
			files = append(files, getFilePaths(path.Join(directoryPath, file.Name()), uploadedFiles)...)
		} else {
			filePath := path.Join(directoryPath, file.Name())

			if fileShouldBeUploaded(filePath, uploadedFiles) {
				files = append(files, filePath)
			}
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

func uploadFile(filePath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse, lock *sync.RWMutex, uploadedFiles *uploadedfiles.UploadedFiles, wg *sync.WaitGroup) {
	uploadResponse := fileuploader.UploadFile(filePath, encryptionPassphrase, authorizationInfo, bucketID)

	if uploadResponse.StatusCode == 200 {
		writeUploadedFileToMap(lock, uploadedFiles, filePath, uploadResponse.FileID)
	} else {
		log.Printf("The uploading of the file %s returned a status code of %d", filePath, uploadResponse.StatusCode)
	}

	wg.Done()
}

func writeUploadedFileToMap(lock *sync.RWMutex, uploadedFiles *uploadedfiles.UploadedFiles, filePath, fileID string) {
	(*lock).Lock()
	defer (*lock).Unlock()
	log.Printf("Adding %s to uploadedFiles. FileId: %s\n", filePath, fileID)
	(*uploadedFiles)[filePath] = uploadedfiles.UploadedFileInfo{LastUploadedTime: time.Now(), FileID: fileID}
}
