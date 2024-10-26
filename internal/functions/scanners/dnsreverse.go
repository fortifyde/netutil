package scanners

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

func PerformDNSReverseLookup(ctx context.Context, hostfile, dnsLookupFile string, outputFunc func(format string, a ...interface{})) error {
	file, err := os.Open(hostfile)
	if err != nil {
		logger.Error("Failed to open hostfile for DNS lookup: %v", err)
		return err
	}
	defer file.Close()

	// Create/truncate the output file
	outFile, err := os.Create(dnsLookupFile)
	if err != nil {
		logger.Error("Failed to create DNS lookup output file: %v", err)
		return err
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(file)
	var ips []string
	for scanner.Scan() {
		line := scanner.Text()
		ip := strings.TrimSpace(line)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("Error reading hostfile: %v", err)
		return err
	}

	outputFunc("[blue]Starting DNS reverse lookup for %d IPs...[-]\n", len(ips))

	for _, ip := range ips {
		if ctx.Err() == context.Canceled {
			outputFunc("[yellow]DNS Reverse Lookup canceled by user.[-]\n")
			return fmt.Errorf("DNS Reverse Lookup canceled")
		}

		outputFunc("[green]Looking up: %s[-]\n", ip)
		cmd := exec.CommandContext(ctx, "dig", "-x", ip, "+short")
		output, err := cmd.Output()
		if err != nil {
			outputFunc("[red]DNS lookup failed for %s: %v[-]\n", ip, err)
			fmt.Fprintf(outFile, "%s\tNo PTR Record (Error: %v)\n", ip, err)
			continue
		}

		domain := strings.TrimSpace(string(output))
		if domain != "" {
			outputFunc("[green]Found: %s -> %s[-]\n", ip, domain)
			fmt.Fprintf(outFile, "%s\t%s\n", ip, domain)
		} else {
			outputFunc("[yellow]No PTR record found for %s[-]\n", ip)
			fmt.Fprintf(outFile, "%s\tNo PTR Record\n", ip)
		}
	}

	outputFunc("[blue]DNS Reverse Lookup completed and saved to %s[-]\n", dnsLookupFile)
	return nil
}
