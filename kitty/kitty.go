package kitty

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type KittyInstance interface {
	SaveSessionTo(path string) error
}

type KittyClient struct {
	Socket string
}

var (
	kittyPath     string
	kittyPathErr  error
	kittyPathOnce sync.Once
)

func ResolveKittyPath() (string, error) {
	kittyPathOnce.Do(func() {
		if p := os.Getenv("KITTY_PATH"); p != "" {
			kittyPath = p
			return
		}
		kittyPath, kittyPathErr = exec.LookPath("kitty")
		if kittyPathErr != nil {
			kittyPathErr = fmt.Errorf(
				"kitty binary not found: set KITTY_PATH env var or ensure kitty is in PATH: %w",
				kittyPathErr,
			)
		}
	})
	return kittyPath, kittyPathErr
}

func NewKittyClient(socket string) *KittyClient {
	return &KittyClient{
		Socket: socket,
	}
}

func (k *KittyClient) SaveSessionTo(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("session path must be absolute, got %q", path)
	}

	if strings.ContainsAny(path, " \t") {
		return fmt.Errorf("session path must not contain whitespace: %q", path)
	}

	args := []string{"@"}

	if k.Socket != "" {
		args = append(args, "--to", k.Socket)
	}
	args = append(args, "action", "save_as_session", "--save-only", path)

	cmd := exec.Command("kitty", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Failed to save session", "error", err, "output", strings.TrimSpace(string(output)))
		return err
	}

	return nil
}
