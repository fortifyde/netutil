package scanners

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

// conducts an enhanced Nmap discovery scan with service version detection
func PerformNmapScan(hostfilePath, selectedInterface, vlanID, nmapOutputPath string) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = selectedInterface + "." + vlanID
	} else {
		interfaceFlag = selectedInterface
	}
	cmd := exec.Command("nmap",
		"-PE", "-PP", "-PM", // ICMP probes
		"-PS22,135-139,445,80,443,5060,2000,3389,53,88,389,636,3268,123", // TCP SYN probes
		"-PU53,161",         // UDP probes
		"-R",                // Reverse-resolve IP addresses
		"--top-ports", "10", // Scan most common ports
		"-sV",                       // Service version detection
		"-O",                        // OS detection
		"--script=smb-os-discovery", // SMB OS discovery script
		"-iL", hostfilePath,
		"-e", interfaceFlag,
		"-oA", strings.TrimSuffix(nmapOutputPath, filepath.Ext(nmapOutputPath)))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error("Nmap Scan command failed: %v", err)
		return err
	}

	logger.Info("Nmap Discovery Scan completed and saved to %s", nmapOutputPath)
	return nil
}
