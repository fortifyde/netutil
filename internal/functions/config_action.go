package functions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fortifyde/netutil/internal/uiutil" // Adjust the import path as necessary
	"github.com/rivo/tview"
)

type Config struct {
	WorkingDirectory string `json:"working_directory"`
	// Add other configuration parameters here
}

const configFileName = "netutil.json"

func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}

	configDir := filepath.Join(homeDir, ".config")
	configFilePath := filepath.Join(configDir, configFileName)

	// Check if the .config directory exists, create it if it doesn't
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			panic(fmt.Sprintf("Failed to create .config directory: %v", err))
		}
	}

	return configFilePath
}

func LoadConfig() (*Config, error) {
	file, err := os.Open(getConfigFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	file, err := os.Create(getConfigFilePath())
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

func EditWorkingDirectory(app *tview.Application, mainView tview.Primitive) error {
	cfg, err := LoadConfig()
	if err != nil {
		uiutil.ShowError(app, fmt.Sprintf("Failed to load config: %v", err), mainView)
		return err
	}

	cmd := exec.Command("zenity", "--file-selection", "--directory")
	output, err := cmd.Output()
	if err != nil {
		uiutil.ShowError(app, fmt.Sprintf("Failed to select directory: %v", err), mainView)
		return err
	}

	newPath := strings.TrimSpace(string(output))
	if newPath == "" {
		uiutil.ShowError(app, "No directory path selected", mainView)
		return fmt.Errorf("no directory path selected")
	}

	cfg.WorkingDirectory = newPath
	err = SaveConfig(cfg)
	if err != nil {
		uiutil.ShowError(app, fmt.Sprintf("Failed to save config: %v", err), mainView)
		return err
	}

	// Display info message about the updated working directory
	uiutil.ShowTimedMessage(app, fmt.Sprintf("Updated working directory to: %s", cfg.WorkingDirectory), mainView, 3*time.Second)

	return nil
}
