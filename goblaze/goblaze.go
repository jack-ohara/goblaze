package goblaze

import (
	"io/ioutil"
	"log"
	"time"
)

type uploadedFiles map[string]time.Time

func UploadDirectories(directories ...string) {
	for _, directoryPath := range directories {
		for _, filePath := range getFilesNames(directoryPath) {

		}
	}
}

func getFilePaths(directoryPath string) []string {
	var files []string

	allFiles, err := ioutil.ReadDir(directoryPath)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range allFiles {
		if file.IsDir() {
			files = append(files, getFilesPaths(directoryPath+file.Name())...)
		} else {
			files = append(files, directoryPath+file.Name())
		}
	}

	return files
}
