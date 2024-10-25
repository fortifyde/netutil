package pkg

type AnalysisResult struct {
	Name   string      `json:"name"`
	Output interface{} `json:"output"`
}

type NetworkTopology struct {
	IP          string `json:"ip"`
	PacketCount int    `json:"packet_count"`
}

type ProtocolDistribution struct {
	Protocol   string `json:"protocol"`
	FrameCount int    `json:"frame_count"`
}

type VLANID struct {
	VLANID string `json:"vlan_id"`
}

type WeakSSLTLS struct {
	SourceIP         string `json:"source_ip"`
	DestinationIP    string `json:"destination_ip"`
	TLSHandshakeCS   string `json:"tls_handshake_ciphersuite"`
	TLSRecordVersion string `json:"tls_record_version"`
}

type OpenPort struct {
	SourceIP string `json:"source_ip"`
	DestIP   string `json:"dest_ip"`
	Port     string `json:"port"`
}

type BroadcastTraffic struct {
	Protocol string `json:"protocol"`
	SourceIP string `json:"source_ip"`
}

type ICMPTraffic struct {
	SourceIP string `json:"source_ip"`
	DestIP   string `json:"dest_ip"`
	ICMPType string `json:"icmp_type"`
}

type DNSServer struct {
	SourceIP string `json:"source_ip"`
}

type SMBUsage struct {
	SourceIP string `json:"source_ip"`
	DestIP   string `json:"dest_ip"`
}

type UnusualProtocol struct {
	Protocol    string `json:"protocol"`
	PacketCount int    `json:"packet_count"`
}

type PotentialDomainController struct {
	SourceIP string `json:"source_ip"`
	Protocol string `json:"protocol"`
}

type STPRootBridge struct {
	VLANID   string `json:"vlan_id"`
	RootMAC  string `json:"root.mac"`
	RootCost string `json:"root.cost"`
}

type SSDPTraffic struct {
	SourceIP      string `json:"source_ip"`
	HTTPUserAgent string `json:"http_user_agent"`
}

type UnencryptedProtocol struct {
	Protocol    string `json:"protocol"`
	PacketCount int    `json:"packet_count"`
}
