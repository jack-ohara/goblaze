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
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
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

func GetConfiguration() ConfigurationValues {
	_, err := os.Stat(getConfigFilePath())

	if os.IsNotExist(err) {
		log.Fatal("Please call goblaze config first")
	}

	configFileData, err := ioutil.ReadFile(getConfigFilePath())

	if err != nil {
		log.Fatal(err)
	}

	configValues := ConfigurationValues{}

	err = json.Unmarshal(configFileData, &configValues)

	if err != nil {
		log.Fatal(err)
	}

	return configValues
}

func SetupConfigFile() {
	keyID := promptForValue("Please enter the key ID: ", false)
	applicationKey := promptForValue("Please enter the application key: ", false)
	bucketID := promptForValue("Please enter the ID of the backblaze bucket to use: ", false)
	encryptionPassphrase := promptForValue("Please enter a passphrase to be used for encryption: ", true)

	if encryptionPassphrase != promptForValue("Please re-enter your passphrase: ", true) {
		log.Fatal("The two passwords do not match")
	}

	configValues := ConfigurationValues{
		EncryptionPassphrase: encryptionPassphrase,
		KeyID:                keyID,
		ApplicationKey:       applicationKey,
		BucketID:             bucketID,
	}

	writeConfigValuesToFile(configValues)
}

func promptForValue(promptText string, secret bool) string {
	fmt.Print(promptText)

	if secret {
		byteValue, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()

		if err != nil {
			log.Fatal(err)
		}

		return string(byteValue)
	}

	reader := bufio.NewReader(os.Stdin)

	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\r\n", "")
	text = strings.ReplaceAll(text, "\n", "")

	return text
}

func getConfigFilePath() string {
	return path.Join(GetConfigDirectory(), "appConfig.json")
}

func writeConfigValuesToFile(configValues ConfigurationValues) {
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
