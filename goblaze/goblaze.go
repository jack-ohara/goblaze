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

type FileWriteMode int

const (
	// Do not overwrite if a file with the same name already exists
	DoNotOverwrite FileWriteMode = iota
	// Overwrite the file if the version on backblaze is more recent than the existing version
	OverwriteOldFiles
	// Overwrite every file on disk with its respective file from backblaze
	AlwaysOverwrite
)

// DownloadOptions defines the configuration of the download you want to perform
type DownloadOptions struct {
	// DirectoryName is the path of the directory to download from backblaze
	DirectoryName string
	// TargetDirectory is the location that the downloaded files will be written to, with respect to the WriteMode.
	// The downloaded files will be written to a directory with the same name as the last directory in the DirectoryName
	TargetDirectory string
	// WriteMode sets the preference of how the downloaded files will be written to disk
	WriteMode FileWriteMode
}

func UploadDirectory(directoryPath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedFiles := uploadedfiles.GetUploadedFiles()
	lock := sync.RWMutex{}

	var wg sync.WaitGroup
	allFiles := getFilePaths(strings.ReplaceAll(directoryPath, "\\", "/"), uploadedFiles)

	numberOfRequests := 0

	const maxNumberOfRequests = 15

	for i := 0; i < len(allFiles); {
		if numberOfRequests <= maxNumberOfRequests {
			wg.Add(1)

			onCompletion := func() {
				wg.Done()

				numberOfRequests--
			}

			go uploadFile(allFiles[i], encryptionPassphrase, bucketID, authorizationInfo, &lock, &uploadedFiles, &onCompletion)

			numberOfRequests++
			i++
		}
	}

	wg.Wait()
	uploadedfiles.WriteUploadedFiles(uploadedFiles)
}

func DownloadDirectory(options DownloadOptions, decryptionPassphrase string, authorizationInfo accountauthorization.AuthorizeAccountResponse) {
	uploadedfiles := uploadedfiles.GetUploadedFiles()

	var wg sync.WaitGroup

	for fileName, uploadedFileInfo := range uploadedfiles {
		if strings.HasPrefix(fileName, options.DirectoryName) && fileShouldBeDownloaded(fileName, &uploadedFileInfo, &options) {
			wg.Add(1)

			go downloadFileAndWriteToDisk(uploadedFileInfo.FileID, decryptionPassphrase, authorizationInfo, uploadedFileInfo.LargeFile, &wg, &options)
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

func uploadFile(filePath, encryptionPassphrase, bucketID string, authorizationInfo accountauthorization.AuthorizeAccountResponse, lock *sync.RWMutex, uploadedFiles *uploadedfiles.UploadedFiles, onCompletionFunc *func()) {
	log.Println("Starting upload of file ", filePath)

	uploadResponse := fileuploader.UploadFile(filePath, encryptionPassphrase, bucketID, authorizationInfo)

	if uploadResponse.StatusCode == 200 {
		log.Println("Successfully uploaded file ", filePath)

		writeUploadedFileToMap(lock, uploadedFiles, filePath, uploadResponse.FileID)
	} else {
		log.Printf("The uploading of the file %s returned a status code of %d\n", filePath, uploadResponse.StatusCode)
	}

	(*onCompletionFunc)()
}

func writeUploadedFileToMap(lock *sync.RWMutex, uploadedFiles *uploadedfiles.UploadedFiles, filePath, fileID string) {
	(*lock).Lock()
	defer (*lock).Unlock()

	log.Printf("Adding %s to uploadedFiles. FileId: %s\n", filePath, fileID)

	(*uploadedFiles)[filePath] = uploadedfiles.UploadedFileInfo{LastUploadedTime: time.Now(), FileID: fileID}

	uploadedfiles.WriteUploadedFiles(*uploadedFiles)
}

func fileShouldBeDownloaded(fileName string, uploadedFileInfo *uploadedfiles.UploadedFileInfo, options *DownloadOptions) bool {
	switch options.WriteMode {
	case AlwaysOverwrite:
		return true
	case OverwriteOldFiles:
		targetFile := getTargetFileName(fileName, options.TargetDirectory, options.DirectoryName)

		fileInfo, err := os.Stat(targetFile)

		if os.IsNotExist(err) {
			return true
		}

		return uploadedFileInfo.LastUploadedTime.After(fileInfo.ModTime().Local())
	case DoNotOverwrite:
		targetFile := getTargetFileName(fileName, options.TargetDirectory, options.DirectoryName)

		_, err := os.Stat(targetFile)

		return os.IsNotExist(err)
	}

	log.Fatalf("The supplied WriteMode option '%d' is not recognised", options.WriteMode)

	return false
}

func downloadFileAndWriteToDisk(fileID, decryptionPassphrase string, authorizationInfo accountauthorization.AuthorizeAccountResponse, largeFile bool, wg *sync.WaitGroup, options *DownloadOptions) {
	downloadResponse := filedownloader.DownloadFileById(fileID, decryptionPassphrase, authorizationInfo, largeFile)

	if downloadResponse.StatusCode != 200 || len(downloadResponse.FileContent) == 0 {
		log.Printf("Something went wrong with the download for file with ID %s. Aborting the write to disk\n", fileID)

		return
	}

	targetFile := getTargetFileName(downloadResponse.FileName, options.TargetDirectory, options.DirectoryName)

	lastSlashIndex := strings.LastIndexByte(targetFile, byte('/'))
	containingDirectory := targetFile[:lastSlashIndex]

	err := os.MkdirAll(containingDirectory, os.ModePerm)

	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(targetFile, downloadResponse.FileContent, os.ModePerm)

	if err != nil {
		log.Println(err)
	}

	(*wg).Done()
}

func getTargetFileName(uploadedFileName, targetDirectory, directoryToDownload string) string {
	directoryToDownload = strings.TrimPrefix(directoryToDownload, "/")
	directoryToDownload = strings.TrimPrefix(directoryToDownload, "\\")

	uploadedFileName = strings.TrimPrefix(uploadedFileName, "/")
	uploadedFileName = strings.TrimPrefix(uploadedFileName, "\\")

	uploadedFileName = strings.TrimPrefix(uploadedFileName, directoryToDownload[:strings.LastIndex(directoryToDownload, string(os.PathSeparator))])

	return path.Join(targetDirectory, uploadedFileName)
}
