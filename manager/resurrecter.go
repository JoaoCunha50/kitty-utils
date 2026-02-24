package manager

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/JoaoCunha50/kitty-utils/config"
	"github.com/JoaoCunha50/kitty-utils/kitty"
	"github.com/JoaoCunha50/kitty-utils/models"
)

type Resurrecter struct {
	Kitty      kitty.KittyInstance
	configDir  string
	outputFile string
}

type ResurrecterInterface interface {
	Listen()
	SaveSession() error
	formatToConfig([]models.OSWindow)
}

func NewResurrecter(kitty kitty.KittyInstance) *Resurrecter {
	configDir, err := config.ResolveKittyConfigDir()
	if err != nil {
		configDir = ""
	}

	return &Resurrecter{
		Kitty:      kitty,
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
		_, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
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

	state, err := r.Kitty.GetState()
	if err != nil {
		return err
	}

	content := r.formatToConfig(state)
	return os.WriteFile(outputFile, []byte(content), 0644)
}

func (r *Resurrecter) formatToConfig(windows []models.OSWindow) string {
	var buffer bytes.Buffer

	for _, osWindow := range windows {
		buffer.WriteString("new_os_window\n")
		buffer.WriteString("os_window_state normal\n\n")

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
