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

// ReadConfig reads the configuration from the config file.
func ReadConfig() (*Config, error) {
	configPath := filepath.Join(".config", "netutil", configFileName)
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file does not exist
			return &Config{
				WorkingDirectory:  getDefaultWorkingDirectory(),
				NetworkInterfaces: make(map[string]InterfaceState),
				DefaultRoute:      "",
			}, nil
		}
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &cfg, nil
}

// WriteConfig writes the configuration to the config file.
func WriteConfig(cfg *Config) error {
	configDir := filepath.Join(".config", "netutil")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, configFileName)
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config to file: %v", err)
	}

	return nil
}

// getDefaultWorkingDirectory returns the default working directory path.
func getDefaultWorkingDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Critical("Failed to get user home directory: %v", err)
		os.Exit(1)
	}
	return filepath.Join(homeDir, "netutil_working_dir")
}

func EditWorkingDirectory(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	cfg, err := ReadConfig()
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
	err = WriteConfig(cfg)
	if err != nil {
		uiutil.ShowError(app, pages, fmt.Sprintf("Failed to save config: %v", err), mainView, nil)
		return err
	}

	uiutil.ShowTimedMessage(app, pages, fmt.Sprintf("Updated working directory to: %s", cfg.WorkingDirectory), mainView, 3*time.Second)
	logger.Info("Updated working directory to: %s", cfg.WorkingDirectory)

	return nil
}
