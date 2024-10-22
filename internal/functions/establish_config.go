package functions

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/rivo/tview"
)

func ReadWriteConfig(app *tview.Application, primitive tview.Primitive) (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	if cfg.WorkingDirectory == "" {
		logger.Info("No working directory set, prompting user to select one")
		exec.Command("zenity", "--info", "--title=File Structure Configuration", "--text=Please select a base working directory.").Run()

		cfg.WorkingDirectory, err = selectWorkingDirectory()
		if err != nil {
			logger.Error("Failed to select working directory: %v", err)
			return nil, fmt.Errorf("failed to select working directory: %v", err)
		}

		err = SaveConfig(cfg)
		if err != nil {
			logger.Error("Failed to save config: %v", err)
			return nil, fmt.Errorf("failed to save config: %v", err)
		}
	}

	logger.Info("Using working directory: %s", cfg.WorkingDirectory)
	return cfg, nil
}

func selectWorkingDirectory() (string, error) {
	// check if zenity is available
	_, err := exec.LookPath("zenity")
	if err == nil {
		cmd := exec.Command("zenity", "--file-selection", "--directory")
		output, err := cmd.Output()
		if err == nil {
			path := strings.TrimSpace(string(output))
			if path != "" {
				logger.Info("Selected directory: %s", path)
				return path, nil
			}
		}
	}

	// if fail, prompt user manually
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter the directory path: ")
		path, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		path = strings.TrimSpace(path)
		if path != "" {
			logger.Info("Selected directory: %s", path)
			return path, nil
		}
		fmt.Println("A directory path must be selected.")
	}
}
