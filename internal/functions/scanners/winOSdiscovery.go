package scanners

import (
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// performs a quick Windows OS discovery and outputs results to windowsDiscoveryFile.
func PerformWindowsOSDiscovery(pingScanFile, selectedInterface, vlanID, windowsDiscoveryFile string) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = selectedInterface + "." + vlanID
	} else {
		interfaceFlag = selectedInterface
	}
	cmd := exec.Command("nmap", "-Pn", "-n", "-p445", "--script=smb-os-discovery", "-oG", windowsDiscoveryFile, "-iL", pingScanFile, "-e", interfaceFlag)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("Windows OS Discovery command failed: %v", err)
		return err
	}

	if err := os.WriteFile(windowsDiscoveryFile, output, 0644); err != nil {
		logger.Error("Failed to write Windows OS Discovery output to file: %v", err)
		return err
	}
	logger.Info("Windows OS Discovery completed and saved to %s", windowsDiscoveryFile)
	return nil
}
