package functions

import (
	"net"
	"strings"
)

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
	return strings.Contains(name, ".") || strings.Contains(name, ":")
}
