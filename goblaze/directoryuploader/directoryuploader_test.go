package directoryuploader_test

import (
	"testing"

	"github.com/jack-ohara/goblaze/goblaze/directoryuploader"
)

func TestUploadDirectory(t *testing.T) {
	directoryuploader.UploadDirectories("/home/jack/Documents/Backup-Test/Dir1", "/home/jack/Documents/Backup-Test/Dir2", "/home/jack/Documents/Backup-Test/Dir3")
}
