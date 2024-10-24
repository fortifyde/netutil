package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

// detectInterfaceForIPRange attempts to automatically detect the appropriate interface or VLAN for the given IP range.
func DetectInterfaceForIPRange(ipRange string) (interfaceName string, vlanID string, err error) {
	// Parse the IP range to get the network
	_, network, err := net.ParseCIDR(ipRange)
	if err != nil {
		logger.Warning("Invalid IP range format: %v", err)
		return "", "", fmt.Errorf("invalid IP range format")
	}

	interfaces, err := GetEthernetInterfaces()
	if err != nil {
		logger.Warning("Failed to get Ethernet interfaces: %v", err)
		return "", "", err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logger.Warning("Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if network.Contains(ip) {
				// Check if the interface is a VLAN subinterface
				if strings.Contains(iface.Name, ".") {
					parts := strings.Split(iface.Name, ".")
					if len(parts) == 2 {
						return parts[0], parts[1], nil
					}
				}
				return iface.Name, "", nil
			}
		}
	}

	logger.Warning("No matching interface found for IP range %s", ipRange)
	return "", "", fmt.Errorf("no matching interface found for IP range")
}
