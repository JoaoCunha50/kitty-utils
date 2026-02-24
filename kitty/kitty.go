package kitty

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
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
	if k.Socket != "" {
		if _, err := os.Stat(k.Socket); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return []models.OSWindow{}, errors.New("socket does not exist: " + k.Socket)
			}
			return []models.OSWindow{}, err
		}
	}

	cmd := exec.Command("kitty", "@", "ls")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to get state", "error", err)
		return []models.OSWindow{}, err
	}

	var windows []models.OSWindow
	if err := json.Unmarshal(output, &windows); err != nil {
		slog.Error("Failed to parse kitty state", "error", err)
		return []models.OSWindow{}, err
	}

	return windows, nil
}
