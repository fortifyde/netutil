package functions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// handles loading and saving of config file
// if file doesn't exist, returns empty config struct
// uses zenity to select working directory with manual fallback

type Config struct {
	WorkingDirectory  string                    `json:"working_directory"`
	NetworkInterfaces map[string]InterfaceState `json:"network_interfaces"`
	DefaultRoute      string                    `json:"default_route"`
}

type InterfaceState struct {
	Status        string              `json:"status"`
	IPAddress     string              `json:"ip_address"`
	SubnetMask    string              `json:"subnet_mask"`
	LinkState     string              `json:"link_state"`
	Subinterfaces []SubinterfaceState `json:"subinterfaces"`
}

type SubinterfaceState struct {
	Name       string `json:"name"`
	IPAddress  string `json:"ip_address"`
	SubnetMask string `json:"subnet_mask"`
}

const configFileName = "netutil.json"

func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}

	configDir := filepath.Join(homeDir, ".config")
	configFilePath := filepath.Join(configDir, configFileName)

	// check if exists, create if not
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

	logger.Info("Config loaded successfully")
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
	err = encoder.Encode(config)
	if err != nil {
		return err
	}

	logger.Info("Config saved successfully")
	return encoder.Encode(config)
}

func EditWorkingDirectory(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	cfg, err := LoadConfig()
	if err != nil {
		uiutil.ShowError(app, pages, fmt.Sprintf("Failed to load config: %v", err), mainView, nil)
		return err
	}

	cmd := exec.Command("zenity", "--file-selection", "--directory")
	output, err := cmd.Output()
	if err != nil {
		uiutil.ShowError(app, pages, fmt.Sprintf("Failed to select directory: %v", err), mainView, nil)
		return err
	}

	newPath := strings.TrimSpace(string(output))
	if newPath == "" {
		uiutil.ShowError(app, pages, "No directory path selected", mainView, nil)
		return fmt.Errorf("no directory path selected")
	}

	cfg.WorkingDirectory = newPath
	err = SaveConfig(cfg)
	if err != nil {
		uiutil.ShowError(app, pages, fmt.Sprintf("Failed to save config: %v", err), mainView, nil)
		return err
	}

	uiutil.ShowTimedMessage(app, pages, fmt.Sprintf("Updated working directory to: %s", cfg.WorkingDirectory), mainView, 3*time.Second)
	logger.Info("Updated working directory to: %s", cfg.WorkingDirectory)

	return nil
}
