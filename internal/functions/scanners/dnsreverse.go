package scanners

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

func PerformDNSReverseLookup(hostfile, dnsLookupFile string) error {
	file, err := os.Open(hostfile)
	if err != nil {
		logger.Error("Failed to open hostfile for DNS lookup: %v", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ips []string
	for scanner.Scan() {
		line := scanner.Text()
		// Assuming the hostfile has one IP per line
		ip := strings.TrimSpace(line)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("Error reading hostfile: %v", err)
		return err
	}

	var result strings.Builder
	for _, ip := range ips {
		cmd := exec.Command("dig", "-x", ip, "+short")
		output, err := cmd.Output()
		if err != nil {
			logger.Warning("DNS lookup failed for %s: %v", ip, err)
			continue
		}
		domain := strings.TrimSpace(string(output))
		if domain != "" {
			result.WriteString(fmt.Sprintf("%s\t%s\n", ip, domain))
		} else {
			result.WriteString(fmt.Sprintf("%s\tNo PTR Record\n", ip))
		}
	}

	if err := os.WriteFile(dnsLookupFile, []byte(result.String()), 0644); err != nil {
		logger.Error("Failed to write DNS Reverse Lookup output to file: %v", err)
		return err
	}
	logger.Info("DNS Reverse Lookup completed and saved to %s", dnsLookupFile)
	return nil
}
