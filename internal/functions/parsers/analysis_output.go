package parsers

import (
	"strings"
)

// define signature for parser functions
type ParserFunc func([]string) (interface{}, error)

// return a map of name to parser functions
func GetParserRegistry() map[string]ParserFunc {
	return map[string]ParserFunc{
		"Network Topology": func(lines []string) (interface{}, error) {
			return parseNetworkTopology(lines)
		},
		"VLAN IDs": func(lines []string) (interface{}, error) {
			return parseVLANIDs(lines)
		},
		"Protocol Distribution": func(lines []string) (interface{}, error) {
			return parseProtocolDistribution(lines)
		},
		"Unencrypted Protocols": func(lines []string) (interface{}, error) {
			return parseUnencryptedProtocols(lines)
		},
		"Unusual Protocols": func(lines []string) (interface{}, error) {
			return parseUnusualProtocols(lines)
		},
		"Weak SSLTLS Versions": func(lines []string) (interface{}, error) {
			return parseWeakSSLTLS(lines)
		},
		"Open Ports": func(lines []string) (interface{}, error) {
			return parseOpenPorts(lines)
		},
		"Broadcast Traffic": func(lines []string) (interface{}, error) {
			return parseBroadcastTraffic(lines)
		},
		"ICMP Traffic": func(lines []string) (interface{}, error) {
			return parseICMPTraffic(lines)
		},
		"DNS Servers": func(lines []string) (interface{}, error) {
			return parseDNSServers(lines)
		},
		"SMB Usage": func(lines []string) (interface{}, error) {
			return parseSMBUsage(lines)
		},
		"Potential Domain Controllers": func(lines []string) (interface{}, error) {
			return parsePotentialDomainControllers(lines)
		},
		"STP Root Bridges": func(lines []string) (interface{}, error) {
			return parseSTPRootBridges(lines)
		},
		"SSDP Traffic": func(lines []string) (interface{}, error) {
			return parseSSDPTraffic(lines)
		},
	}
}

// parse raw output from tshark based on analysis type
func ParseAnalysisOutput(name string, data string) (interface{}, error) {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	registry := GetParserRegistry()

	if parser, exists := registry[name]; exists {
		return parser(lines)
	}

	return lines, nil
}
