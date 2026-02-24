package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/JoaoCunha50/kitty-utils/config"
	"github.com/JoaoCunha50/kitty-utils/kitty"
	"github.com/JoaoCunha50/kitty-utils/manager"
)

func main() {
	socket := flag.String("socket", "", "Kitty socket path (e.g., unix:@mykitty)")
	flag.Parse()

	if *socket == "" {
		fmt.Println("Error: -socket flag is required")
		flag.Usage()
		os.Exit(1)
	}

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

	kittyClient := kitty.NewKittyClient(*socket)
	resurrecter := manager.NewResurrecter(kittyClient)
	resurrecter.Listen()
}
