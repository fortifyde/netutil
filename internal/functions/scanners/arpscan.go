package scanners

import (
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// conducts an ARP scan over the specified IP range and outputs results to arpScanFile.
func PerformARPscan(ipRange, selectedInterface, vlanID, arpScanFile string) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = "--interface=" + selectedInterface + "." + vlanID
	} else {
		interfaceFlag = "--interface=" + selectedInterface
	}

	cmd := exec.Command("arp-scan", interfaceFlag, ipRange)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("ARP Scan command failed: %v", err)
		return err
	}

	if err := os.WriteFile(arpScanFile, output, 0644); err != nil {
		logger.Error("Failed to write ARP Scan output to file: %v", err)
		return err
	}
	logger.Info("ARP Scan completed and saved to %s", arpScanFile)
	return nil
}
