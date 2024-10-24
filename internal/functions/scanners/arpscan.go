package scanners

import (
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// performARPSscan conducts an ARP scan over the specified IP range and outputs results to arpScanFile.
func PerformARPscan(ipRange, arpScanFile string) error {
	cmd := exec.Command("arp-scan", "--interface=eth0", ipRange) // Adjust interface as needed
	output, err := cmd.Output()
	if err != nil {
		logger.Error("ARP Scan command failed: %v", err)
		return err
	}

	// Write output to file
	if err := os.WriteFile(arpScanFile, output, 0644); err != nil {
		logger.Error("Failed to write ARP Scan output to file: %v", err)
		return err
	}
	logger.Info("ARP Scan completed and saved to %s", arpScanFile)
	return nil
}
