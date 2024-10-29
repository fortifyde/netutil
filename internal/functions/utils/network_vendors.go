package utils

import (
	"strings"
)

// Common networking equipment vendors and their identifiers
var networkVendors = map[string][]string{
	"cisco": {
		"ios",
		"nexus",
		"catalyst",
		"aironet",
		"meraki",
		"cisco systems",
	},
	"juniper": {
		"junos",
		"juniper networks",
		"srx",
		"ex series",
		"qfx",
	},
	"arista": {
		"eos",
		"arista networks",
		"dcs-",
	},
	"hpe": { // Hewlett Packard Enterprise
		"procurve",
		"aruba",
		"hp networking",
		"comware",
		"provision",
	},
	"extreme": {
		"extremexos",
		"extreme networks",
		"summit",
		"black diamond",
	},
}

// Common ports and services for network devices
var networkPorts = map[int]string{
	22:   "ssh",
	23:   "telnet",
	80:   "http",
	443:  "https",
	161:  "snmp",
	162:  "snmptrap",
	514:  "syslog",
	830:  "netconf-ssh",
	3389: "management",
	8443: "https-alt",
}

func isNetworkDevice(host HostInfo) (bool, string) {
	// Check OS/Vendor information from Nmap
	vendorLower := strings.ToLower(host.OS.Vendor)
	for vendor, identifiers := range networkVendors {
		for _, id := range identifiers {
			if strings.Contains(vendorLower, id) {
				return true, vendor
			}
		}
	}

	// Check service banners and product information
	for _, port := range host.Ports {
		serviceLower := strings.ToLower(port.Service.Product)
		for vendor, identifiers := range networkVendors {
			for _, id := range identifiers {
				if strings.Contains(serviceLower, id) {
					return true, vendor
				}
			}
		}
	}

	// Check for typical network device port combinations
	networkPortCount := 0
	hasSSH := false
	hasSNMP := false

	for _, port := range host.Ports {
		if _, isNetworkPort := networkPorts[port.PortID]; isNetworkPort {
			networkPortCount++
		}
		if port.PortID == 22 {
			hasSSH = true
		}
		if port.PortID == 161 {
			hasSNMP = true
		}
	}

	// If device has multiple network management ports including SSH and SNMP,
	// it's likely a network device
	if networkPortCount >= 3 && hasSSH && hasSNMP {
		return true, "unknown"
	}

	return false, ""
}
