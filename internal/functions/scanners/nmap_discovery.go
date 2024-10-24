package scanners

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

// performNmapScan conducts an Nmap discovery scan using the hostfile and saves output in all formats.
func PerformNmapScan(hostfilePath, nmapOutputPath string) error {
	// Define the Nmap scan command with desired output formats
	cmd := exec.Command("nmap", "-sP", "-iL", hostfilePath, "-oA", strings.TrimSuffix(nmapOutputPath, filepath.Ext(nmapOutputPath)))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error("Nmap Scan command failed: %v", err)
		return err
	}

	logger.Info("Nmap Discovery Scan completed and saved to %s", nmapOutputPath)
	return nil
}
