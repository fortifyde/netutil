package parsers

import (
	"strconv"
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

// parse output for Network Topology analysis
// extract IP addresses and packet counts
func parseNetworkTopology(lines []string) ([]pkg.NetworkTopology, error) {
	var topology []pkg.NetworkTopology
	parse := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Start parsing after the header separator
		if strings.HasPrefix(line, "===") {
			parse = true
			continue
		}

		if !parse || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		ip := fields[0]
		packetCount, err := strconv.Atoi(fields[1])
		if err != nil {
			continue // Skip lines with invalid packet counts
		}

		topology = append(topology, pkg.NetworkTopology{
			IP:          ip,
			PacketCount: packetCount,
		})
	}

	return topology, nil
}
