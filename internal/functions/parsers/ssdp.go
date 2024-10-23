package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseSSDPTraffic(lines []string) ([]pkg.SSDPTraffic, error) {
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

	var ssdpList []pkg.SSDPTraffic
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ssdpList = append(ssdpList, pkg.SSDPTraffic{
				SourceIP:      fields[0],
				HTTPUserAgent: fields[1],
			})
		}
	}
	return ssdpList, nil
}
