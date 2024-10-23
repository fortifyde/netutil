package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseUnusualProtocols(lines []string) ([]pkg.UnusualProtocol, error) {
	protocolMap := make(map[string]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		protocol := fields[0]

		if protocol == "Protocol" || protocol == "No" {
			continue
		}

		protocolMap[protocol]++
	}

	var unusualProtocols []pkg.UnusualProtocol
	for protocol, count := range protocolMap {
		unusualProtocols = append(unusualProtocols, pkg.UnusualProtocol{
			Protocol:    protocol,
			PacketCount: count,
		})
	}

	return unusualProtocols, nil
}
