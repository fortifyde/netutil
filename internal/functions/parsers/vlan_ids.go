package parsers

import (
	"strings"

	"github.com/fortifyde/netutil/internal/pkg"
)

// extract VLAN IDs from lines
func parseVLANIDs(lines []string) ([]pkg.VLANID, error) {
	var vlanIDs []pkg.VLANID
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		vlanID := fields[0]

		if vlanID == "VLAN" || strings.Contains(vlanID, "Running") {
			continue
		}

		vlanIDs = append(vlanIDs, pkg.VLANID{
			VLANID: vlanID,
		})
	}

	return vlanIDs, nil
}
