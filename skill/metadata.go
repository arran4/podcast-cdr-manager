package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Metadata struct {
	Source      string    `json:"source"`
	Revision    string    `json:"revision"`
	InstallTime time.Time `json:"install_time"`
	Digest      string    `json:"digest"`
}

func ReadMetadata(skillDir string) (*Metadata, error) {
	path := filepath.Join(skillDir, ".skill-metadata.json")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var md Metadata
	if err := json.NewDecoder(f).Decode(&md); err != nil {
		return nil, err
	}
	return &md, nil
}

func WriteMetadata(skillDir string, md *Metadata) error {
	path := filepath.Join(skillDir, ".skill-metadata.json")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(md)
}
