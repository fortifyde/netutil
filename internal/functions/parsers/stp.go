package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

func parseSTPRootBridges(lines []string) ([]pkg.STPRootBridge, error) {
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
	var stpBridges []pkg.STPRootBridge
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			stpBridges = append(stpBridges, pkg.STPRootBridge{
				VLANID:   fields[0],
				RootMAC:  fields[1],
				RootCost: fields[2],
			})
		}
	}
	return stpBridges, nil
}
