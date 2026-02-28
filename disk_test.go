package podcast_cdr_manager

import (
	"testing"
)

func TestCreateDiskFilename(t *testing.T) {
	// Let's call createDiskFilename with increasing values of i
	// to see if we can trigger the panic.
	for i := 0; i < 10000; i++ {
		_ = createDiskFilename(i)
	}
}

func TestCreateDiskIsoName(t *testing.T) {
	for i := 0; i < 10000; i++ {
		_ = createDiskIsoName(i)
	}
}
