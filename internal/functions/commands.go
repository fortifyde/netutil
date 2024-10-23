package functions

type AnalysisCommand struct {
	Name string
	Args []string
}

// get tshark analysis commands
func GetAnalysisCommands(captureFile string) []AnalysisCommand {
	return []AnalysisCommand{
		{
			Name: "Network Topology",
			Args: []string{"-r", captureFile, "-q", "-z", "endpoints,ip"},
		},
		{
			Name: "VLAN IDs",
			Args: []string{"-r", captureFile, "-Y", "vlan", "-T", "fields", "-e", "vlan.id"},
		},
		{
			Name: "Protocol Distribution",
			Args: []string{"-r", captureFile, "-q", "-z", "io,phs"},
		},
		{
			Name: "Unencrypted Protocols",
			Args: []string{"-r", captureFile, "-Y", "http || ftp || telnet || pop || imap", "-T", "fields", "-e", "_ws.col.protocol"},
		},
		{
			Name: "Open Ports",
			Args: []string{"-r", captureFile, "-Y", "tcp.flags.syn == 1 && tcp.flags.ack == 0", "-T", "fields", "-e", "ip.src", "-e", "ip.dst", "-e", "tcp.dstport"},
		},
		{
			Name: "Broadcast Traffic",
			Args: []string{"-r", captureFile, "-Y", "eth.dst == ff:ff:ff:ff:ff:ff", "-T", "fields", "-e", "eth.type", "-e", "ip.src"},
		},
		{
			Name: "Potential Domain Controllers",
			Args: []string{"-r", captureFile, "-Y", "domain_controller_protocol_filter", "-T", "fields", "-e", "ip.src", "-e", "protocol"},
		},
		{
			Name: "STP Root Bridges",
			Args: []string{"-r", captureFile, "-Y", "stp", "-T", "fields", "-e", "vlan.id", "-e", "stp.root.hw", "-e", "stp.root.cost"},
		},
		{
			Name: "SSDP Traffic",
			Args: []string{"-r", captureFile, "-Y", "ssdp", "-T", "fields", "-e", "ip.src", "-e", "http.user_agent"},
		},
		{
			Name: "Unusual Protocols",
			Args: []string{"-r", captureFile, "-Y", "!http && !ftp && !telnet && !pop && !imap && !ssh && !dns", "-T", "fields", "-e", "_ws.col.protocol"},
		},
		{
			Name: "Weak SSLTLS Versions",
			Args: []string{"-r", captureFile, "-Y", "tls.record.version < 0x0300 || tls.handshake.ciphersuite in {0x0001, 0x0002, 0x0003}", "-T", "fields", "-e", "ip.src", "-e", "ip.dst", "-e", "tls.handshake.ciphersuite", "-e", "tls.record.version"},
		},
	}
}
