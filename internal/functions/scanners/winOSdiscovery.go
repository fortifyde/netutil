package scanners

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// performs a quick Windows OS discovery and outputs results to windowsDiscoveryFile.
func PerformWindowsOSDiscovery(ctx context.Context, pingScanFile, selectedInterface, vlanID, windowsDiscoveryFile string, outputFunc func(format string, a ...interface{})) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = selectedInterface + "." + vlanID
	} else {
		interfaceFlag = selectedInterface
	}

	cmd := exec.CommandContext(ctx, "nmap", "-Pn", "-n", "-p445", "--script=smb-os-discovery", "-oG", windowsDiscoveryFile, "-iL", pingScanFile, "-e", interfaceFlag)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to get stdout pipe for Windows OS Discovery: %v", err)
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to get stderr pipe for Windows OS Discovery: %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start Windows OS Discovery command: %v", err)
		return err
	}

	// Create channels to signal completion of output processing
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	// Process stdout
	go func() {
		defer close(stdoutDone)
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			outputFunc("[green]Windows OS Discovery: %s[-]\n", scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			outputFunc("[red]Windows OS Discovery Error: %s[-]\n", scanner.Text())
		}
	}()

	// Wait for both output streams to complete
	<-stdoutDone
	<-stderrDone

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			outputFunc("[yellow]Windows OS Discovery canceled by user.[-]")
			return fmt.Errorf("windows OS Discovery canceled")
		}
		logger.Error("Windows OS Discovery command failed: %v", err)
		return err
	}

	// Write the output to the file
	// Since Nmap outputs to multiple files with -oA, we'll assume the XML file needs to be handled
	// Adjust as necessary based on actual requirements

	outputFunc("[blue]Windows OS Discovery completed and saved to %s[-]", windowsDiscoveryFile)
	return nil
}
