package main

import (
	"log"
	"os"

	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	authorizationInfo := goblaze.GetAccountAuthorization(os.Getenv("KEY_ID"), os.Getenv("APPLICATION_ID"))

	fileuploader.UploadFile("/home/jack/helloworld.txt", authorizationInfo, os.Getenv("BUCKET_ID"))
}
