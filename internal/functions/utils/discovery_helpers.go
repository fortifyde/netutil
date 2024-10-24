package utils

import (
	"bufio"
	"os"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

// createHostfile aggregates all IP addresses from various scan outputs into a single hostfile.
func CreateHostfile(scanFiles []string, hostfilePath string) error {
	hostSet := make(map[string]struct{})

	for _, filePath := range scanFiles {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Warning("Failed to open scan file %s: %v", filePath, err)
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			ip := ExtractIPFromLine(line)
			if ip != "" {
				hostSet[ip] = struct{}{}
			}
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			logger.Warning("Error scanning file %s: %v", filePath, err)
		}
	}

	var hosts []string
	for ip := range hostSet {
		hosts = append(hosts, ip)
	}

	hostContent := strings.Join(hosts, "\n")
	if err := os.WriteFile(hostfilePath, []byte(hostContent), 0644); err != nil {
		logger.Error("Failed to write hostfile: %v", err)
		return err
	}
	logger.Info("Hostfile created at %s with %d hosts", hostfilePath, len(hosts))
	return nil
}
