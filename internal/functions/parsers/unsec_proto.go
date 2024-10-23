package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseUnencryptedProtocols(lines []string) ([]pkg.UnencryptedProtocol, error) {
	protocolMap := make(map[string]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 {
			protocol := fields[0]

			if protocol == "Protocol" || protocol == "No" || protocol == "Running" {
				continue
			}

			protocolMap[protocol]++
		}
	}

	var unencryptedProtocols []pkg.UnencryptedProtocol
	for protocol, count := range protocolMap {
		unencryptedProtocols = append(unencryptedProtocols, pkg.UnencryptedProtocol{
			Protocol:    protocol,
			PacketCount: count,
		})
	}

	return unencryptedProtocols, nil
}
