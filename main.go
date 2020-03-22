package main

import (
	"github.com/jack-ohara/goblaze/goblaze"
	"github.com/jack-ohara/goblaze/goblaze/fileuploader"
)

func main() {
	authorizationInfo := goblaze.GetAccountAuthorization("7f3a2586df8e", "0000ac10cd67b60d1e4333a8cfff77aee02297a008")

	fileuploader.UploadFile("/home/jack/helloworld.txt", authorizationInfo, "67cf33ca2205c8667d0f081e")
}
