package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parsePotentialDomainControllers(lines []string) ([]pkg.PotentialDomainController, error) {
	protocolMap := make(map[string]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 {
			protocol := fields[0]

			if protocol == "Protocol" || protocol == "Running" {
				continue
			}

			protocolMap[protocol]++
		}
	}

	var domainControllers []pkg.PotentialDomainController
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			domainControllers = append(domainControllers, pkg.PotentialDomainController{
				SourceIP: fields[0],
				Protocol: fields[1],
			})
		}
	}
	return domainControllers, nil
}
