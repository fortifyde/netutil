package scanners

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// PerformARPscan conducts an ARP scan over the specified IP range and outputs results to arpScanFile.
// It streams command outputs to the provided outputFunc.
func PerformARPscan(ctx context.Context, ipRange, selectedInterface, vlanID, arpScanFile string, outputFunc func(format string, a ...interface{})) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = "--interface=" + selectedInterface + "." + vlanID
	} else {
		interfaceFlag = "--interface=" + selectedInterface
	}

	cmd := exec.CommandContext(ctx, "arp-scan", interfaceFlag, ipRange)

	// Create/truncate the output file
	file, err := os.Create(arpScanFile)
	if err != nil {
		logger.Error("Failed to create output file for ARP Scan: %v", err)
		return err
	}
	defer file.Close()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to get stdout pipe for ARP Scan: %v", err)
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to get stderr pipe for ARP Scan: %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start ARP Scan command: %v", err)
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
			line := scanner.Text()
			outputFunc("[green]ARP Scan: %s[-]\n", line)
			fmt.Fprintln(file, line)
		}
	}()

	// Process stderr
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			outputFunc("[red]ARP Scan Error: %s[-]\n", scanner.Text())
		}
	}()

	// Wait for both output streams to complete
	<-stdoutDone
	<-stderrDone

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			outputFunc("[yellow]ARP Scan canceled by user.[-]\n")
			return fmt.Errorf("ARP Scan canceled")
		}
		logger.Error("ARP Scan command failed: %v", err)
		return err
	}

	outputFunc("[blue]ARP Scan completed and saved to %s[-]\n", arpScanFile)
	return nil
}
