package kitty

import (
	"encoding/json"
	"log/slog"
	"os/exec"

	"github.com/JoaoCunha50/kitty-utils/models"
)

type KittyInstance interface {
	GetState() ([]models.OSWindow, error)
}

type KittyClient struct {
	Socket string
}

func NewKittyClient(socket string) *KittyClient {
	return &KittyClient{
		Socket: socket,
	}
}

func (k *KittyClient) GetState() ([]models.OSWindow, error) {
	var cmd *exec.Cmd
	args := []string{"@", "ls"}
	if k.Socket != "" {
		args = []string{"@", "--to", k.Socket, "ls"}
	}
	cmd = exec.Command("kitty", args...)

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
