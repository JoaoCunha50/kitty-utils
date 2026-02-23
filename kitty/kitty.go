package kitty

import (
	"encoding/json"
	"log/slog"
	"os/exec"

	"../models"
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
