package scanners

import (
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// performPingScan conducts an asynchronous ping scan using fping and outputs results to pingScanFile.
func PerformPingScan(ipRange, pingScanFile string) error {
	cmd := exec.Command("fping", "-a", "-g", ipRange)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("Ping Scan command failed: %v", err)
		return err
	}

	// Write output to file
	if err := os.WriteFile(pingScanFile, output, 0644); err != nil {
		logger.Error("Failed to write Ping Scan output to file: %v", err)
		return err
	}
	logger.Info("Ping Scan completed and saved to %s", pingScanFile)
	return nil
}
