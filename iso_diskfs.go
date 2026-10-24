package podcast_cdr_manager

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
)

const isoBlockSize int64 = 2048

// diskfsISOImageWriter creates ISO9660 with Rock Ridge and Joliet extensions.
// Files are staged once and handed to go-diskfs as its workspace, avoiding an
// extra copy of every MP3 during filesystem assembly.
type diskfsISOImageWriter struct {
	stagedISOFiles
}

func newDiskfsISOImageWriter() (isoImageWriter, error) {
	dir, err := os.MkdirTemp("", "podcast-cdr-diskfs-")
	if err != nil {
		return nil, fmt.Errorf("create diskfs ISO workspace: %w", err)
	}
	return &diskfsISOImageWriter{stagedISOFiles: stagedISOFiles{stagingDir: dir}}, nil
}

// isoPayloadAndRoom uses a conservative sparse-image provision. Finalize()
// records the real sector count in the primary volume descriptor; we copy only
// that many sectors to the final output, not the unused provisioned capacity.
func (w *diskfsISOImageWriter) isoPayloadAndRoom() (int64, error) {
	var total, count int64
	err := filepath.Walk(w.stagingDir, func(p string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() {
			if info.Size() < 0 || info.Size() > (1<<62)-total {
				return fmt.Errorf("ISO payload is too large")
			}
			total += (info.Size()+isoBlockSize-1)/isoBlockSize*isoBlockSize
			count++
		}
		return nil
	})
	if err != nil { return 0, fmt.Errorf("size staged files: %w", err) }
	// Reserved sparse headroom for directory records, path tables, Joliet,
	// Rock Ridge continuations and volume descriptors.
	const baseline int64 = 4 << 20
	if count > (1<<62-total-baseline)/(16<<10) {
		return 0, fmt.Errorf("too many ISO files")
	}
	return total + baseline + count*(16<<10), nil
}

func (w *diskfsISOImageWriter) WriteTo(destination io.Writer, volumeName string) error {
	if w.stagingDir == "" {
		return fmt.Errorf("diskfs ISO writer has been cleaned up")
	}
	provision, err := w.isoPayloadAndRoom()
	if err != nil { return err }
	imgDir, err := os.MkdirTemp("", "podcast-cdr-diskfs-image-")
	if err != nil { return fmt.Errorf("create ISO workspace: %w", err) }
	defer func(){ _ = os.RemoveAll(imgDir) }()
	imgPath := filepath.Join(imgDir, "image.iso")

	imageDisk, err := diskfs.Create(imgPath, provision, diskfs.SectorSizeDefault)
	if err != nil { return fmt.Errorf("create diskfs image: %w", err) }
	defer func(){ _ = imageDisk.Close() }()
	imageDisk.LogicalBlocksize = isoBlockSize
	fsGeneric, err := imageDisk.CreateFilesystem(disk.FilesystemSpec{
		Partition: 0,
		FSType: filesystem.TypeISO9660,
		WorkDir: w.stagingDir,
	})
	if err != nil { return fmt.Errorf("create ISO filesystem: %w", err) }
	fs, ok := fsGeneric.(*iso9660.FileSystem)
	if !ok { _ = fsGeneric.Close(); return fmt.Errorf("unexpected ISO filesystem implementation") }
	defer func(){ _ = fs.Close() }()
	if err := fs.Finalize(iso9660.FinalizeOptions{
		RockRidge: true,
		Joliet: true,
		VolumeIdentifier: volumeName,
	}); err != nil { return fmt.Errorf("finalize diskfs image: %w", err) }

	// The PVD stores the *used* volume length at byte 80 (LSB, 32-bit).
	// The initialized image may have substantial spare provisioned sectors.
	imageFile, err := os.Open(imgPath)
	if err != nil { return fmt.Errorf("open created ISO: %w", err) }
	defer func(){ _ = imageFile.Close() }()
	var pvd [2048]byte
	if _, err := imageFile.ReadAt(pvd[:], 16*isoBlockSize); err != nil {
		return fmt.Errorf("read ISO primary volume descriptor: %w", err)
	}
	if pvd[0] != 1 || string(pvd[1:6]) != "CD001" ||
		binary.LittleEndian.Uint16(pvd[128:130]) != uint16(isoBlockSize) {
		return fmt.Errorf("go-diskfs produced an invalid primary volume descriptor")
	}
	sectors := binary.LittleEndian.Uint32(pvd[80:84])
	usedBytes := int64(sectors)*isoBlockSize
	stat, err := imageFile.Stat()
	if err != nil { return err }
	if usedBytes <= 0 || usedBytes > stat.Size() {
		return fmt.Errorf("invalid ISO used extent %d (provisioned %d)", usedBytes, stat.Size())
	}
	if _, err := imageFile.Seek(0, io.SeekStart); err != nil { return err }
	if _, err := io.CopyN(destination, imageFile, usedBytes); err != nil {
		return fmt.Errorf("copy finalized ISO: %w", err)
	}
	return nil
}
