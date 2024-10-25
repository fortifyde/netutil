package scanners

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/fortifyde/netutil/internal/logger"
)

// PerformPingScan conducts an asynchronous ping scan using fping and outputs results to pingScanFile.
// It streams command outputs to the provided outputFunc.
func PerformPingScan(ctx context.Context, ipRange, selectedInterface, vlanID, pingScanFile string, outputFunc func(format string, a ...interface{})) error {
	var interfaceFlag string
	if vlanID != "" {
		interfaceFlag = selectedInterface + "." + vlanID
	} else {
		interfaceFlag = selectedInterface
	}

	cmd := exec.CommandContext(ctx, "fping", "-a", "-g", ipRange, "-I", interfaceFlag, "-q")

	// Create/truncate the output file
	file, err := os.Create(pingScanFile)
	if err != nil {
		logger.Error("Failed to create output file for Ping Scan: %v", err)
		return err
	}
	defer file.Close()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to get stdout pipe for Ping Scan: %v", err)
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to get stderr pipe for Ping Scan: %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start Ping Scan command: %v", err)
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
			outputFunc("[green]Ping Scan: %s[-]\n", line)
			fmt.Fprintln(file, line)
		}
	}()

	// Stream stderr
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			outputFunc("[yellow]Ping Scan Warning: %s[-]\n", scanner.Text())
		}
	}()

	// Wait for both output streams to complete
	<-stdoutDone
	<-stderrDone

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			outputFunc("[yellow]Ping Scan canceled by user.[-]")
			return fmt.Errorf("ping Scan canceled")
		}
		logger.Error("Ping Scan command failed: %v", err)
		return err
	}

	outputFunc("[blue]Ping Scan completed and saved to %s[-]", pingScanFile)
	return nil
}
