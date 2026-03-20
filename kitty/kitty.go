package kitty

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	"github.com/JoaoCunha50/kitty-utils/models"
)

type KittyInstance interface {
	GetState() ([]models.OSWindow, error)
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

func (k *KittyClient) GetState() ([]models.OSWindow, error) {
	binPath, err := ResolveKittyPath()
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	args := []string{"@", "ls"}
	if k.Socket != "" {
		args = []string{"@", "--to", k.Socket, "ls"}
	}
	cmd = exec.Command(binPath, args...)

	output, err := cmd.Output()
	if err != nil {
		slog.Error("kitty ls", "output", string(output))
		slog.Error("Failed to get state", "error", err)
		return nil, err
	}

	var windows []models.OSWindow
	if err := json.Unmarshal(output, &windows); err != nil {
		slog.Error("Failed to parse kitty state", "error", err)
		return nil, err
	}

	return windows, nil
}
