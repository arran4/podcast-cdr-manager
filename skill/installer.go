package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Installer struct{}

func NewInstaller() *Installer {
	return &Installer{}
}

func (i *Installer) Install(sourceStr string, target Target, force bool) error {
	src, err := ParseSource(sourceStr)
	if err != nil {
		return err
	}

	parts := strings.Split(sourceStr, "/")
	skillName := parts[len(parts)-1]
	if strings.HasPrefix(sourceStr, "./") || strings.HasPrefix(sourceStr, "../") || strings.HasPrefix(sourceStr, "/") {
		skillName = filepath.Base(sourceStr)
	}

	if skillName == "" || skillName == "." || skillName == ".." || strings.ContainsAny(skillName, `/\\`) {
		return fmt.Errorf("invalid skill name: %q", skillName)
	}

	dest := target.InstallPath(skillName)

	if _, err := os.Stat(dest); err == nil && !force {
		return fmt.Errorf("skill '%s' is already installed at %s. Use --force to overwrite", skillName, dest)
	}

	tempDest := dest + ".tmp"
	_ = os.RemoveAll(tempDest)
	if err := os.MkdirAll(tempDest, 0755); err != nil {
		return err
	}

	defer os.RemoveAll(tempDest)

	revision, err := src.Fetch(tempDest)
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}

	md := &Metadata{
		Source:      sourceStr,
		Revision:    revision,
		InstallTime: time.Now(),
		Digest:      hashDir(tempDest),
	}

	if err := WriteMetadata(tempDest, md); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	_ = os.RemoveAll(dest)
	if err := os.Rename(tempDest, dest); err != nil {
		return fmt.Errorf("failed to finalize installation: %w", err)
	}

	fmt.Printf("Successfully installed skill '%s' to %s\n", skillName, dest)
	return nil
}

func (i *Installer) Update(skillName string, target Target, force bool) error {
	if skillName == "" || skillName == "." || skillName == ".." || strings.ContainsAny(skillName, `/\\`) {
		return fmt.Errorf("invalid skill name: %q", skillName)
	}

	dest := target.InstallPath(skillName)

	md, err := ReadMetadata(dest)
	if err != nil {
		return fmt.Errorf("could not read metadata for skill '%s': %w", skillName, err)
	}

	currentDigest := hashDir(dest)
	if currentDigest != md.Digest && !force {
		return fmt.Errorf("installed skill has local modifications; rerun with --force to replace")
	}

	src, err := ParseSource(md.Source)
	if err != nil {
		return err
	}

	tempDest := dest + ".tmp"
	os.RemoveAll(tempDest)
	if err := os.MkdirAll(tempDest, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(tempDest)

	revision, err := src.Fetch(tempDest)
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}

	if revision == md.Revision && currentDigest == md.Digest && !force {
		fmt.Printf("Skill '%s' is already up to date.\n", skillName)
		return nil
	}

	newMd := &Metadata{
		Source:      md.Source,
		Revision:    revision,
		InstallTime: time.Now(),
		Digest:      hashDir(tempDest),
	}

	if err := WriteMetadata(tempDest, newMd); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	os.RemoveAll(dest)
	if err := os.Rename(tempDest, dest); err != nil {
		return fmt.Errorf("failed to finalize update: %w", err)
	}

	fmt.Printf("Successfully updated skill '%s'\n", skillName)
	return nil
}

func (i *Installer) Remove(skillName string, target Target) error {
	if skillName == "" || skillName == "." || skillName == ".." || strings.ContainsAny(skillName, `/\\`) {
		return fmt.Errorf("invalid skill name: %q", skillName)
	}
	dest := target.InstallPath(skillName)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return fmt.Errorf("skill '%s' not found", skillName)
	}

	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	fmt.Printf("Successfully removed skill '%s'\n", skillName)
	return nil
}
