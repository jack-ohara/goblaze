package goblaze

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/filedownloader"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
	"github.com/jack-ohara/goblaze/goblaze/uploadedfiles"
)

func UploadDirectories(directoryPath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedFiles := uploadedfiles.GetUploadedFiles()
	lock := sync.RWMutex{}

	var wg sync.WaitGroup

	for _, filePath := range getFilePaths(directoryPath, uploadedFiles) {
		wg.Add(1)

		go uploadFile(filePath, encryptionPassphrase, bucketID, authorizationInfo, &lock, &uploadedFiles, &wg)
	}

	wg.Wait()
	uploadedfiles.WriteUploadedFiles(uploadedFiles)
}

func DownloadDirectory(directoryName, decryptionPassphrase string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedfiles := uploadedfiles.GetUploadedFiles()

	var wg sync.WaitGroup

	for fileName, uploadedFileInfo := range uploadedfiles {
		if strings.HasPrefix(fileName, directoryName) {
			wg.Add(1)

			go downloadFileAndWriteToDisk(uploadedFileInfo.FileID, decryptionPassphrase, authorizationInfo, uploadedFileInfo.LargeFile, &wg)
		}
	}

	wg.Wait()
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
	uploadResponse := fileuploader.UploadFile(filePath, encryptionPassphrase, bucketID, authorizationInfo)

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

func downloadFileAndWriteToDisk(fileID, decryptionPassphrase string, authorizationInfo accountauthorization.AuthorizeAccountResponse, largeFile bool, wg *sync.WaitGroup) {
	downloadResponse := filedownloader.DownloadFileById(fileID, decryptionPassphrase, authorizationInfo, largeFile)

	fileName := "/" + downloadResponse.FileName

	if downloadResponse.StatusCode != 200 || len(downloadResponse.FileContent) == 0 {
		log.Printf("Something went wrong with the download for file %s. Aborting the write to disk\n", fileName)

		return
	}

	lastSlashIndex := strings.LastIndexByte(fileName, byte('/'))
	containingDirectory := fileName[:lastSlashIndex]

	err := os.MkdirAll(containingDirectory, os.ModePerm)

	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(fileName, downloadResponse.FileContent, os.ModePerm)

	if err != nil {
		log.Println(err)
	}

	(*wg).Done()
}
