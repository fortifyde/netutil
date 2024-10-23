package parsers

import (
	"strconv"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/pkg"
)

func parseProtocolDistribution(lines []string) ([]pkg.ProtocolDistribution, error) {
	var protocolDist []pkg.ProtocolDistribution
	parse := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// start parsing after first "===" line
		if strings.HasPrefix(line, "===") {
			if !parse {
				parse = true
				continue
			} else {
				break
			}
		}

		if !parse || line == "" {
			continue
		}

		if !strings.Contains(line, "frames:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		protocol := fields[0]
		var frameCount int

		for _, field := range fields[1:] {
			if strings.HasPrefix(field, "frames:") {
				frameStr := strings.TrimPrefix(field, "frames:")
				frameStr = strings.ReplaceAll(frameStr, ",", "")
				count, err := strconv.Atoi(frameStr)
				if err != nil {
					logger.Error("Failed to parse frame count for protocol %s: %v", protocol, err)
					frameCount = 0
				} else {
					frameCount = count
				}
				break
			}
		}

		protocolDist = append(protocolDist, pkg.ProtocolDistribution{
			Protocol:   protocol,
			FrameCount: frameCount,
		})
	}

	return protocolDist, nil
}
