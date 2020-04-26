package configuration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type ConfigurationValues struct {
	EncryptionPassphrase string
	KeyID                string
	ApplicationKey       string
	BucketID             string
}

func GetConfigDirectory() string {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	return path.Join(userHomeDir, ".goblaze")
}

func GetUploadedFilesPath() string {
	return path.Join(GetConfigDirectory(), "uploadedFiles.json")
}

func SetupConfigFile() {
	keyID := promptForValue("Please enter the key ID: ")
	applicationKey := promptForValue("Please enter the application key: ")
	bucketID := promptForValue("Please enter the ID of the backblaze bucket to use: ")
	encryptionPassphrase := promptForValue("Please enter a passphrase to be used for encryption: ")

	configValues := ConfigurationValues{
		EncryptionPassphrase: encryptionPassphrase,
		KeyID:                keyID,
		ApplicationKey:       applicationKey,
		BucketID:             bucketID,
	}

	jsonData, err := json.MarshalIndent(configValues, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	configFilePath := getConfigFilePath()

	if _, err := os.Stat(GetConfigDirectory()); os.IsNotExist(err) {
		os.MkdirAll(GetConfigDirectory(), os.ModePerm)
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		file, err := os.Create(configFilePath)

		if err != nil {
			log.Fatal(err)
		}

		file.Close()
	}

	err = ioutil.WriteFile(getConfigFilePath(), jsonData, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}

func promptForValue(promptText string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(promptText)
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "\r\n", "")

	return text
}

func getConfigFilePath() string {
	return path.Join(GetConfigDirectory(), "appConfig.json")
}
