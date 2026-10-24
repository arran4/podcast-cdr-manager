package podcast_cdr_manager

import (
	"fmt"
	"io"
	"os"

	// Default image creation is implemented in Go using go-diskfs.
)

// isoImageWriter is the narrow image-mastering boundary. Alternative backends
// can be implemented independently of download and profile-management logic.
type isoImageWriter interface {
	AddFile(io.Reader, string) error
	AddLocalFile(string, string) error
	WriteTo(io.Writer, string) error
	Cleanup() error
}

// newISOImageWriter chooses the pure-Go Rock Ridge/Joliet backend by default.
// xorriso remains an optional alternate implementation and independent validator.
func newISOImageWriter() (isoImageWriter, error) {
	switch os.Getenv("PODCAST_CDR_ISO_BACKEND") {
	case "", "diskfs":
		return newDiskfsISOImageWriter()
	case "xorriso":
		return newXorrisoImageWriter()
	default:
		return nil, fmt.Errorf("unknown ISO backend %q (choose diskfs or xorriso)", os.Getenv("PODCAST_CDR_ISO_BACKEND"))
	}
}
