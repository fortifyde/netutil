package utils

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
)

// GetEthernetInterfaces retrieves all wired Ethernet interfaces, excluding wireless and subinterfaces.
func GetEthernetInterfaces() ([]net.Interface, error) {
	allInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ethernetInterfaces []net.Interface

	for _, iface := range allInterfaces {
		if isWiredEthernetInterface(iface.Name) && !isSubinterface(iface.Name) {
			ethernetInterfaces = append(ethernetInterfaces, iface)
		}
	}

	return ethernetInterfaces, nil
}

// isWiredEthernetInterface checks if the interface name corresponds to a wired Ethernet interface.
func isWiredEthernetInterface(name string) bool {
	wiredPrefixes := []string{
		"eth", "en", "em", "eno", "enp", "ens", "enx",
	}

	for _, prefix := range wiredPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

// isSubinterface determines if the given interface name is a subinterface.
func isSubinterface(name string) bool {
	re := regexp.MustCompile(`\.\d+(@.*)?$`)
	return re.MatchString(name)
}

// IsValidInterface validates whether the provided interface name exists and is a valid Ethernet interface.
func IsValidInterface(ifaceName string) bool {
	// Check if the interface exists
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return false
	}

	// Further validation: ensure it's a wired Ethernet interface
	return isWiredEthernetInterface(iface.Name) && !isSubinterface(iface.Name)
}

// GetSubinterfaces retrieves all subinterfaces for a given parent interface.
func GetSubinterfaces(ifaceName string) ([]string, error) {
	cmd := exec.Command("ip", "-o", "link", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get subinterfaces: %v", err)
	}

	var subinterfaces []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		fullSubName := strings.TrimRight(fields[1], ":")
		if isSubinterfaceOf(fullSubName, ifaceName) {
			subName := strings.Split(fullSubName, "@")[0]
			subinterfaces = append(subinterfaces, subName)
		}
	}

	return subinterfaces, nil
}

// isSubinterfaceOf checks if a subinterface belongs to the specified parent interface.
func isSubinterfaceOf(subName, parentName string) bool {
	pattern := fmt.Sprintf(`^%s\.\d+(@%s)?$`, regexp.QuoteMeta(parentName), regexp.QuoteMeta(parentName))
	match, _ := regexp.MatchString(pattern, subName)
	return match
}

func GetInterfaceStatus(name string) (string, error) {
	logger.Info("Getting status for interface: %s", name)
	cmd := exec.Command("ip", "link", "show", name)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("Failed to get status for interface %s: %v", name, err)
		return "", err
	}

	if strings.Contains(string(output), "state UP") {
		return "up", nil
	}
	return "down", nil
}
