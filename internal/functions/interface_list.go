package functions

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// retrieve all ethernet interfaces
// filter out wireless interfaces and subinterfaces
// separate function to retrieve subinterfaces of a given interface
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

func isSubinterface(name string) bool {
	re := regexp.MustCompile(`\.\d+(@.*)?$`)
	return re.MatchString(name)
}

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

func isSubinterfaceOf(subName, parentName string) bool {
	pattern := fmt.Sprintf(`^%s\.\d+(@%s)?$`, regexp.QuoteMeta(parentName), regexp.QuoteMeta(parentName))
	match, _ := regexp.MatchString(pattern, subName)
	return match
}
