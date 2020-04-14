package fileuploader_test

import (
	"os"
	"testing"

	"github.com/jack-ohara/goblaze/goblaze/accountauthorization"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
	"github.com/joho/godotenv"
)

func TestLargeFileUpload(t *testing.T) {
	err := godotenv.Load()

	if err != nil {
		t.Fatal(err)
	}

	authorizationInfo := accountauthorization.GetAccountAuthorization(os.Getenv("KEY_ID"), os.Getenv("APPLICATION_ID"))

	response := fileuploader.UploadFile("/home/jack/Documents/Backup-Test/Coppice_06-06-2018.zip", os.Getenv("ENCRYPTION_PASSPHRASE"), os.Getenv("BUCKET_ID"), authorizationInfo)

	t.Logf("%+v\n", response)
}
