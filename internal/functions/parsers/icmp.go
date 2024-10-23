package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseICMPTraffic(lines []string) ([]pkg.ICMPTraffic, error) {
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

	var icmpList []pkg.ICMPTraffic
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			icmpList = append(icmpList, pkg.ICMPTraffic{
				SourceIP: fields[0],
				DestIP:   fields[1],
				ICMPType: fields[2],
			})
		}
	}
	return icmpList, nil
}
