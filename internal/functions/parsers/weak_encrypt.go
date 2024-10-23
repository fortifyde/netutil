package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseWeakSSLTLS(lines []string) ([]pkg.WeakSSLTLS, error) {
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

	var sslVersions []pkg.WeakSSLTLS
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			sslVersions = append(sslVersions, pkg.WeakSSLTLS{
				SourceIP:         fields[0],
				DestinationIP:    fields[1],
				TLSHandshakeCS:   fields[2],
				TLSRecordVersion: fields[3],
			})
		}
	}
	return sslVersions, nil
}
