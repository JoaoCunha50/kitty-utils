package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"../../config"
	"../../kitty"
	"../../manager"
)

func main() {
	var handler slog.Handler
	var file *os.File
	var err error
	configDir, err := config.ResolveKittyConfigDir()
	if err != nil {
		configDir = ""
	}

	file, err = os.OpenFile(filepath.Join(configDir, "kitty-resurrecter.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(file, nil)
	}

	slog.SetDefault(slog.New(handler))

	listenOn := os.Getenv("KITTY_LISTEN_ON")
	if listenOn == "" {
		listenOn = "unix:/mykitty"
		_ = os.Setenv("KITTY_LISTEN_ON", listenOn)
	}

	kittyClient := kitty.NewKittyClient(listenOn)
	resurrecter := manager.NewResurrecter(kittyClient)
	resurrecter.Listen()
}
