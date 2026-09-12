package manager

import (
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JoaoCunha50/kitty-utils/config"
	"github.com/JoaoCunha50/kitty-utils/kitty"
)

type Resurrecter struct {
	configDir  string
	outputFile string
	lastSocket string
	mu         sync.Mutex
}

type ResurrecterInterface interface {
	Listen()
	SaveSession() error
}

func NewResurrecter() *Resurrecter {
	configDir, err := config.ResolveKittyConfigDir()
	if err != nil {
		configDir = ""
	}

	return &Resurrecter{
		configDir:  configDir,
		outputFile: filepath.Join(configDir, "kitty-session.conf"),
	}
}

func (r *Resurrecter) Listen() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:11223")
	if err != nil {
		slog.Error("Failed to resolve UDP address", "error", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.Error("Failed to start UDP listener", "error", err)
		return
	}
	defer conn.Close()

	var debounceTimer *time.Timer
	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			slog.Error("Failed to read from UDP", "error", err)
			continue
		}

		socket := strings.TrimSpace(string(buffer[:n]))
		if socket != "" {
			r.mu.Lock()
			r.lastSocket = socket
			r.mu.Unlock()
		}

		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		debounceTimer = time.AfterFunc(1*time.Second, func() {
			if err := r.SaveSession(); err != nil {
				slog.Error("Auto-save failed", "error", err)
			}
		})
	}
}

func (r *Resurrecter) SaveSession() error {
	outputFile, err := config.ExpandPath(r.outputFile)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(outputFile) {
		return fmt.Errorf("could not resolve an absolute session path from %q", r.outputFile)
	}

	r.mu.Lock()
	socket := r.lastSocket
	r.mu.Unlock()

	if socket == "" {
		return fmt.Errorf("no kitty instance registered yet")
	}

	slog.Info("Saving session...", "file", outputFile, "socket", socket)
	return kitty.NewKittyClient(socket).SaveSessionTo(outputFile)
}
