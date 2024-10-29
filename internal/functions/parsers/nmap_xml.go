package parsers

import (
	"encoding/xml"
)

// NmapRun represents the root of the Nmap XML.
type NmapRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts   []Host   `xml:"host"`
}

// Host represents each host entry in Nmap XML.
type Host struct {
	Status     Status       `xml:"status"`
	Address    Address      `xml:"address"`
	HostNames  []HostName   `xml:"hostnames>hostname"`
	Ports      []Port       `xml:"ports>port"`
	OSMatches  []OSMatch    `xml:"os>osmatch"`
	OSClasses  []OSClass    `xml:"os>osclass"`
	HostScript []HostScript `xml:"hostscript>script"`
	Truncated  bool         `xml:"truncated,attr"`
}

/* Additional structs to parse specific XML elements */

type Status struct {
	State string `xml:"state,attr"`
}

type Address struct {
	Addr string `xml:"addr,attr"`
	// AddrType can be "ipv4" or "ipv6"
	AddrType string `xml:"addrtype,attr"`
}

type HostName struct {
	Name     string `xml:"name,attr"`
	HostType string `xml:"type,attr"`
}

type Port struct {
	Protocol string      `xml:"protocol,attr"`
	PortID   int         `xml:"portid,attr"`
	State    PortState   `xml:"state"`
	Service  PortService `xml:"service"`
}

type PortState struct {
	State     string `xml:"state,attr"`
	Reason    string `xml:"reason,attr"`
	ReasonTTL string `xml:"reason_ttl,attr"`
}

type PortService struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
	Extra   string `xml:"extrainfo,attr"`
}

type OS struct {
	OSMatches []OSMatch `xml:"osmatch"`
}

type OSMatch struct {
	Name      string    `xml:"name,attr"`
	Accuracy  int       `xml:"accuracy,attr"`
	OSClasses []OSClass `xml:"osclass"`
}

type OSClass struct {
	Type     string `xml:"type,attr"`
	Vendor   string `xml:"vendor,attr"`
	OSFamily string `xml:"osfamily,attr"`
	OSGen    string `xml:"osgen,attr"`
	TypeEx   string `xml:"type_ex,attr"`
	Accuracy int    `xml:"accuracy,attr"`
}

type HostScript struct {
	ID     string `xml:"id,attr"`
	Output string `xml:"output,attr"`
}
