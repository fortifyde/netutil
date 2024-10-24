package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
)

// check dependencies to ensure required binaries are installed
func CheckDependencies(binaries []string) bool {
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin); err != nil {
			return false
		}
	}
	return true
}

// ensure directory exists and create if it does not
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dirName, err)
	}
	return nil
}

// get the working directory from the configuration
func GetWorkingDirectory() string {
	cfg, err := configuration.LoadConfig()
	if err != nil {
		logger.Critical("Failed to read configuration: %v", err)
		os.Exit(1)
	}
	return cfg.WorkingDirectory
}

// extractIPFromLine extracts the IP address from a given line.
func ExtractIPFromLine(line string) string {
	fields := strings.Fields(line)
	for _, field := range fields {
		if net.ParseIP(field) != nil {
			return field
		}
	}
	return ""
}
