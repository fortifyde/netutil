package scanners

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

// PerformNmapScan conducts an enhanced Nmap discovery scan with service version detection.
// It streams command outputs to the provided outputFunc.
func PerformNmapScan(ctx context.Context, hostfilePath, selectedInterface, vlanID, nmapOutputPath string, outputFunc func(format string, a ...interface{})) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = selectedInterface + "." + vlanID
	} else {
		interfaceFlag = selectedInterface
	}

	cmd := exec.CommandContext(ctx, "nmap",
		"-PE", "-PP", "-PM", // ICMP probes
		"-PS22,135-139,445,80,443,5060,2000,3389,53,88,389,636,3268,123", // TCP SYN probes
		"-PU53,161", // UDP probes
		//"-R",                // Reverse-resolve IP addresses
		"-n",
		"--top-ports", "10", // Scan most common ports
		"-sV",                       // Service version detection
		"-O",                        // OS detection
		"--script=smb-os-discovery", // SMB OS discovery script
		"--min-hostgroup", " 64",
		"--min-parallelism", "32",
		"--host-timeout", "10m",
		"-iL", hostfilePath,
		"-e", interfaceFlag,
		"-oA", strings.TrimSuffix(nmapOutputPath, filepath.Ext(nmapOutputPath)),
	)
	logger.Info("Starting Nmap Discovery Scan with command: %s", cmd.String())

	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to get stdout pipe for Nmap Scan: %v", err)
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to get stderr pipe for Nmap Scan: %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start Nmap Scan command: %v", err)
		return err
	}

	// Stream stdout
	go func() {
		defer close(stdoutDone)
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			outputFunc("[green]Nmap Scan: %s[-]\n", scanner.Text())
		}
	}()

	// Process stderr
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			outputFunc("[red]Nmap Scan Error: %s[-]\n", scanner.Text())
		}
	}()

	// Wait for both output streams to complete
	<-stdoutDone
	<-stderrDone

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			outputFunc("[yellow]Nmap Scan canceled by user.[-]")
			return fmt.Errorf("nmap Scan canceled")
		}
		logger.Error("Nmap Scan command failed: %v", err)
		return err
	}

	// Write the output to the file
	// Since Nmap outputs to multiple files with -oA, we'll assume the XML file needs to be handled
	// Adjust as necessary based on actual requirements
	logger.Info("Nmap Discovery Scan completed and saved to %s", nmapOutputPath)
	outputFunc("[blue]Nmap Discovery Scan completed and saved to %s[-]\n", nmapOutputPath)
	return nil
}
