package manager

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JoaoCunha50/kitty-utils/config"
	"github.com/JoaoCunha50/kitty-utils/kitty"
	"github.com/JoaoCunha50/kitty-utils/models"
)

type Resurrecter struct {
	configDir     string
	outputFile    string
	activeSockets map[string]struct{}
	mu            sync.Mutex
}

type ResurrecterInterface interface {
	Listen()
	SaveSession() error
	formatToConfig([]models.OSWindow)
}

func NewResurrecter() *Resurrecter {
	configDir, err := config.ResolveKittyConfigDir()
	if err != nil {
		configDir = ""
	}

	return &Resurrecter{
		configDir:     configDir,
		outputFile:    filepath.Join(configDir, "kitty-session.conf"),
		activeSockets: make(map[string]struct{}),
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
			r.activeSockets[socket] = struct{}{}
			r.mu.Unlock()
			slog.Info("Registered socket", "socket", socket)
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

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}

	r.mu.Lock()
	sockets := make([]string, 0, len(r.activeSockets))
	for s := range r.activeSockets {
		sockets = append(sockets, s)
	}
	r.mu.Unlock()

	var allWindows []models.OSWindow
	for _, socket := range sockets {
		client := kitty.NewKittyClient(socket)
		windows, err := client.GetState()
		if err != nil {
			r.mu.Lock()
			delete(r.activeSockets, socket)
			r.mu.Unlock()
			slog.Warn("Socket unavailable, removing", "socket", socket, "error", err)
			continue
		}
		allWindows = append(allWindows, windows...)
	}

	if len(allWindows) == 0 {
		return fmt.Errorf("no windows found")
	}

	content := r.formatToConfig(allWindows)
	slog.Info("Saving session...", "file", outputFile, "windows", len(allWindows))
	return os.WriteFile(outputFile, []byte(content), 0644)
}

func (r *Resurrecter) formatToConfig(windows []models.OSWindow) string {
	var buffer bytes.Buffer

	for _, osWindow := range windows {
		if len(windows) > 1 {
			buffer.WriteString("new_os_window\n")
			buffer.WriteString("os_window_state normal\n\n")
		}

		for _, tab := range osWindow.Tabs {
			if len(tab.Windows) == 0 {
				continue
			}

			fmt.Fprintf(&buffer, "new_tab %s\n", tab.Title)
			if tab.Layout != "" {
				fmt.Fprintf(&buffer, "layout %s\n", tab.Layout)
			}

			for _, window := range tab.Windows {
				if window.Cwd != "" {
					fmt.Fprintf(&buffer, "cd %s\n", window.Cwd)
				}
				buffer.WriteString("launch\n")
			}
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}
