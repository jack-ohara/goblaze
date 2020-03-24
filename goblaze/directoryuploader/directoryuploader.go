package directoryuploader

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

type directory struct {
	Name        string
	Directories []directory
	Files       []file
}

type file struct {
	FileName     string
	LastUploaded time.Time
}

func UploadDirectories(directories ...string) {
	directoriesToUpload := []directory{}

	for _, directoryPath := range directories {
		directoriesToUpload = append(directoriesToUpload, getDirectory(directoryPath))
	}
}

func getDirectory(directoryPath string) directory {
	fileInfos, err := ioutil.ReadDir(directoryPath)

	if err != nil {
		log.Fatal(err)
	}

	currentDirectory := directory{Name: directoryPath}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			dir := getDirectory(filepath.Join(directoryPath, fileInfo.Name()))

			currentDirectory.Directories = append(currentDirectory.Directories, dir)
		} else {
			currentDirectory.Files = append(currentDirectory.Files, file{FileName: fileInfo.Name()})
		}
	}

	return currentDirectory
}
