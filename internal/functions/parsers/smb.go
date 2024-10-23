package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseSMBUsage(lines []string) ([]pkg.SMBUsage, error) {
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

	var smbUsages []pkg.SMBUsage
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			smbUsages = append(smbUsages, pkg.SMBUsage{
				SourceIP: fields[0],
				DestIP:   fields[1],
			})
		}
	}
	return smbUsages, nil
}
