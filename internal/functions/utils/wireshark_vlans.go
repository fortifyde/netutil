package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
)

// GetWiresharkVLANs retrieves unique VLAN IDs from Wireshark capture files
func GetWiresharkVLANs() ([]string, error) {
	cfg, err := configuration.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	captureDir := filepath.Join(cfg.WorkingDirectory, "captures")
	vlanMap := make(map[string]bool)

	files, err := os.ReadDir(captureDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read captures directory: %v", err)
	}

	vlanRegex := regexp.MustCompile(`VLAN ID: (\d+)`)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		f, err := os.Open(filepath.Join(captureDir, file.Name()))
		if err != nil {
			logger.Warning("Failed to open capture file %s: %v", file.Name(), err)
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			matches := vlanRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				vlanMap[matches[1]] = true
			}
		}
	}

	var vlanIDs []string
	for vlanID := range vlanMap {
		vlanIDs = append(vlanIDs, vlanID)
	}
	return vlanIDs, nil
}
