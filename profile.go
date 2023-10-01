package podcast_cdr_manager

import (
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Profile struct {
	Version       string
	Name          string
	Subscriptions []*Subscription
	Casts         []*Cast
	Disks         []*Disk
}

func NewProfile(name string) (*Profile, error) {
	if len(name) < 2 {
		return nil, fmt.Errorf("profile name too short")
	}
	if d, _ := filepath.Split(name); len(d) > 0 {
		return nil, fmt.Errorf("profile name must be a valid filename")
	}
	if e := filepath.Ext(name); len(e) > 0 {
		return nil, fmt.Errorf("profile name must not contain a file extension")
	}
	configFilePath, err := xdg.ConfigFile(filepath.Join("podcast-cdr-manager", name+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("config directory: %w", err)
	}
	_, err = os.ReadFile(configFilePath)
	if err == nil {
		return nil, fmt.Errorf("profile already exists")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading profile data: %w", err)
	}

	return &Profile{
		Name:    name,
		Version: "1",
	}, nil
}

func OpenProfile(name string) (*Profile, error) {
	if d, _ := filepath.Split(name); len(d) > 0 {
		return nil, fmt.Errorf("profile name must be a valid filename")
	}
	configFilePath, err := xdg.ConfigFile(filepath.Join("podcast-cdr-manager", name+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("config directory: %w", err)
	}
	b, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading profile data: %w", err)
	}
	var p Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("config structure: %w", err)
	}
	switch p.Version {
	case "1":
	default:
		return nil, fmt.Errorf("unknown version: %w", err)
	}
	p.Name = name
	return &p, nil
}

func (p *Profile) Save() error {
	if d, _ := filepath.Split(p.Name); len(d) > 0 {
		return fmt.Errorf("profile name must be a valid filename")
	}
	configFilePath, err := xdg.ConfigFile(filepath.Join("podcast-cdr-manager", p.Name+".yaml"))
	if err != nil {
		return fmt.Errorf("config directory: %w", err)
	}
	b, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("config structure: %w", err)
	}
	if err := os.WriteFile(configFilePath, b, 0644); err != nil {
		return fmt.Errorf("reading profile data: %w", err)
	}
	return nil
}
