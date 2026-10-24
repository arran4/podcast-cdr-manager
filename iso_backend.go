package podcast_cdr_manager

import (
	"io"

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

func newISOImageWriter() (isoImageWriter, error) {
	return iso9660.NewWriter()
}
