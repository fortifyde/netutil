package scanners

import (
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// performWindowsOSDiscovery performs a quick Windows OS discovery and outputs results to windowsDiscoveryFile.
func PerformWindowsOSDiscovery(ipRange, windowsDiscoveryFile string) error {
	// Using nmap for OS detection as an example
	cmd := exec.Command("nmap", "-O", ipRange)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("Windows OS Discovery command failed: %v", err)
		return err
	}

	// Write output to file
	if err := os.WriteFile(windowsDiscoveryFile, output, 0644); err != nil {
		logger.Error("Failed to write Windows OS Discovery output to file: %v", err)
		return err
	}
	logger.Info("Windows OS Discovery completed and saved to %s", windowsDiscoveryFile)
	return nil
}
