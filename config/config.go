package config

import (
	"os"
	"path/filepath"
)

func ResolveKittyConfigDir() (string, error) {
	if dir := os.Getenv("KITTY_CONFIG_DIRECTORY"); dir != "" {
		return dir, nil
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "kitty"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "kitty"), nil
}

func ExpandPath(file string) (string, error) {
	if file == "" {
		return file, nil
	}
	if filepath.IsAbs(file) {
		return file, nil
	}
	dir, err := ResolveKittyConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, file), nil
}
