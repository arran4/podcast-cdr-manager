package podcast_cdr_manager

import (
	"fmt"
	"io"
	"os"

	"github.com/kdomanski/iso9660"
)

// isoImageWriter is the narrow image-mastering boundary. Alternative backends
// can be implemented independently of download and profile-management logic.
type isoImageWriter interface {
	AddFile(io.Reader, string) error
	AddLocalFile(string, string) error
	WriteTo(io.Writer, string) error
	Cleanup() error
}

// newISOImageWriter selects an ISO writer. The legacy Go writer remains the
// default; the xorriso backend is an explicit opt-in for long names and Joliet.
func newISOImageWriter() (isoImageWriter, error) {
	switch os.Getenv("PODCAST_CDR_ISO_BACKEND") {
	case "", "legacy":
		return iso9660.NewWriter()
	case "xorriso":
		return newXorrisoImageWriter()
	default:
		return nil, fmt.Errorf("unknown ISO backend %q (choose legacy or xorriso)", os.Getenv("PODCAST_CDR_ISO_BACKEND"))
	}
}
