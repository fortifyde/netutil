package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseOpenPorts(lines []string) ([]pkg.OpenPort, error) {
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

	var openPorts []pkg.OpenPort
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			openPorts = append(openPorts, pkg.OpenPort{
				SourceIP: fields[0],
				DestIP:   fields[1],
				Port:     fields[2],
			})
		}
	}
	return openPorts, nil
}
