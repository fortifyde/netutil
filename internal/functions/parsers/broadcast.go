package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseBroadcastTraffic(lines []string) ([]pkg.BroadcastTraffic, error) {
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

	var broadcasts []pkg.BroadcastTraffic
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			broadcasts = append(broadcasts, pkg.BroadcastTraffic{
				Protocol: fields[0],
				SourceIP: fields[1],
			})
		}
	}
	return broadcasts, nil
}
