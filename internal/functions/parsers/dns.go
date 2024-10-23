package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseDNSServers(lines []string) ([]pkg.DNSServer, error) {
	protocolMap := make(map[string]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 {
			protocol := fields[0]

			if protocol == "Running" {
				continue
			}

			protocolMap[protocol]++
		}
	}

	var dnsServers []pkg.DNSServer
	uniqueIPs := make(map[string]bool)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			srcIP := fields[0]
			if !uniqueIPs[srcIP] {
				dnsServers = append(dnsServers, pkg.DNSServer{
					SourceIP: srcIP,
				})
				uniqueIPs[srcIP] = true
			}
		}
	}
	return dnsServers, nil
}
