package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/JoaoCunha50/kitty-utils/config"
	"github.com/JoaoCunha50/kitty-utils/manager"
)

func main() {
	var handler slog.Handler
	var file *os.File
	var err error
	configDir, err := config.ResolveKittyConfigDir()
	if err != nil {
		configDir = ""
		if homeDir, homeErr := os.UserHomeDir(); homeErr == nil && homeDir != "" {
			configDir = filepath.Join(homeDir, ".config", "kitty")
		}
	}

	if configDir == "" {
		configDir = "."
	}

	file, err = os.OpenFile(filepath.Join(configDir, "kitty-utils", "logs", "kitty-resurrecter.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(file, nil)
	}
	defer file.Close()

	slog.SetDefault(slog.New(handler))

	resurrecter := manager.NewResurrecter()
	resurrecter.Listen()
}
