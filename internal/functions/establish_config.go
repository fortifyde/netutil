package functions

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

func ReadWriteConfig(app *tview.Application, primitive tview.Primitive) {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.WorkingDirectory == "" {
		exec.Command("zenity", "--info", "--title=File Structure Configuration", "--text=Please select a base working directory.").Run()

		cfg.WorkingDirectory, err = selectWorkingDirectory()
		if err != nil {
			log.Fatalf("Failed to select working directory: %v", err)
		}

		err = SaveConfig(cfg)
		if err != nil {
			log.Fatalf("Failed to save config: %v", err)
		}
	}

	fmt.Printf("Using working directory: %s\n", cfg.WorkingDirectory)
}

func selectWorkingDirectory() (string, error) {
	// Check if zenity is available
	_, err := exec.LookPath("zenity")
	if err == nil {
		cmd := exec.Command("zenity", "--file-selection", "--directory")
		output, err := cmd.Output()
		if err == nil {
			path := strings.TrimSpace(string(output))
			if path != "" {
				fmt.Printf("Selected directory: %s\n", path)
				return path, nil
			}
		}
	}

	// If zenity is not available or no directory was selected, prompt user manually
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter the directory path: ")
		path, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		path = strings.TrimSpace(path)
		if path != "" {
			fmt.Printf("Selected directory: %s\n", path)
			return path, nil
		}
		fmt.Println("A directory path must be selected.")
	}
}
