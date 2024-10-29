package utils

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fortifyde/netutil/internal/functions/parsers"
	"github.com/fortifyde/netutil/internal/logger"
)

// Category represents a classification category for IP addresses.
type Category string

const (
	CategoryWindowsServer       Category = "Windows_Server"
	CategoryWindowsClient       Category = "Windows_Client"
	CategoryWindowsUnknown      Category = "Windows_Unknown"
	CategoryLinux               Category = "Linux"
	CategoryPrinter             Category = "Printers"
	CategoryUPS                 Category = "UPS"
	CategoryNAS                 Category = "NAS"
	CategoryFirewall            Category = "Firewalls"
	CategoryRouterSwitch        Category = "Routers_Switches"
	CategoryLightsOutManagement Category = "Lights_Out_Management"
	CategoryUnknown             Category = "Unknown"
)

// HostInfo holds parsed information from Nmap XML
type HostInfo struct {
	IP        string
	MAC       string
	MACVendor string
	HostNames []string
	Ports     []Port
	OS        struct {
		Vendor   string
		Family   string
		Gen      string
		Detail   string
		Accuracy int
	}
	Services []Service
}

type Port struct {
	Protocol string
	PortID   int
	State    string
	Service  Service
}

type Service struct {
	Name      string
	Product   string
	Version   string
	ExtraInfo string
}

// Common UPS vendors and their identifiers
var upsVendors = []string{
	"apc",        // APC/Schneider
	"schneider",  // Schneider Electric (APC parent company)
	"eaton",      // Eaton
	"powerware",  // Eaton Powerware
	"emerson",    // Emerson/Vertiv
	"vertiv",     // Vertiv (formerly Emerson)
	"liebert",    // Liebert/Vertiv
	"cyberpower", // CyberPower
	"tripp lite", // Tripp Lite
	"delta",      // Delta
	"riello",     // Riello
}

// Common firewall vendors and their identifiers
var firewallVendors = []string{
	"palo alto",
	"checkpoint",
	"fortinet",
	"fortigate",
	"cisco asa",
	"cisco firepower",
	"sophos",
	"watchguard",
	"sonicwall",
	"pfsense",
	"opnsense",
}

// Common printer vendors and their identifiers
var printerVendors = []string{
	"hp",      // Hewlett-Packard
	"lexmark", // Lexmark
	"canon",   // Canon
	"brother", // Brother
	"epson",   // Epson
	"xerox",   // Xerox
	"ricoh",   // Ricoh
	"kyocera", // Kyocera
	"konica",  // Konica Minolta
	"sharp",   // Sharp
}

// networkDevices stores IPs by network device vendor
var networkDevices = make(map[string][]string)

// Add MAC vendor mapping while keeping existing vendor lists
var macVendors = map[string]Category{
	// Printers (matching existing printerVendors)
	"hewlett packard": CategoryPrinter,
	"hp inc":          CategoryPrinter,
	"lexmark":         CategoryPrinter,
	"canon":           CategoryPrinter,
	"brother":         CategoryPrinter,
	"epson":           CategoryPrinter,
	"xerox":           CategoryPrinter,
	"ricoh":           CategoryPrinter,
	"kyocera":         CategoryPrinter,
	"konica minolta":  CategoryPrinter,
	"sharp":           CategoryPrinter,

	// UPS (matching existing upsVendors)
	"schneider electric": CategoryUPS,
	"apc":                CategoryUPS,
	"eaton":              CategoryUPS,
	"powerware":          CategoryUPS,
	"emerson":            CategoryUPS,
	"vertiv":             CategoryUPS,
	"liebert":            CategoryUPS,
	"cyber power":        CategoryUPS,
	"tripp-lite":         CategoryUPS,
	"delta":              CategoryUPS,
	"riello":             CategoryUPS,

	// Firewalls (matching existing firewallVendors)
	"palo alto":     CategoryFirewall,
	"check point":   CategoryFirewall,
	"fortinet":      CategoryFirewall,
	"cisco systems": CategoryFirewall,
	"sophos":        CategoryFirewall,
	"watchguard":    CategoryFirewall,
	"sonicwall":     CategoryFirewall,
}

// ARPResult holds the parsed information from ARP scan
type ARPResult struct {
	IP     string
	MAC    string
	Vendor string
}

// parseARPScanResults parses the ARP scan file and returns structured data
func parseARPScanResults(arpScanFile string) (map[string]ARPResult, error) {
	results := make(map[string]ARPResult)

	file, err := os.Open(arpScanFile)
	if err != nil {
		logger.Error("Failed to open ARP scan file: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		ip := ExtractIPFromLine(line)
		if ip == "" {
			logger.Debug("No valid IP found in line: %s", line)
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			logger.Debug("Skipping invalid ARP line: %s", line)
			continue
		}

		mac := fields[1]
		vendor := strings.ToLower(strings.Join(fields[2:], " "))

		results[ip] = ARPResult{
			IP:     ip,
			MAC:    mac,
			Vendor: vendor,
		}
		logger.Debug("Found device: IP=%s, MAC=%s, Vendor=%s", ip, mac, vendor)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error reading ARP scan file: %v", err)
		return nil, err
	}

	return results, nil
}

// PerformCategorization processes scan results and categorizes each IP.
func PerformCategorization(hostfilesDir, arpScanFile, pingScanFile, dnsLookupFile, windowsDiscoveryFile, nmapXMLPath string) error {
	logger.Info("Starting host categorization process")

	// Verify input files exist
	logger.Debug("Verifying input files...")
	files := map[string]string{
		"ARP scan":          arpScanFile,
		"Ping scan":         pingScanFile,
		"DNS lookup":        dnsLookupFile,
		"Windows discovery": windowsDiscoveryFile,
		"Nmap XML":          nmapXMLPath,
	}

	for name, path := range files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logger.Error("Required file %s (%s) does not exist", name, path)
			return fmt.Errorf("required file %s (%s) does not exist", name, path)
		}
		logger.Debug("Found %s file: %s", name, path)
	}

	// Verify output directory exists
	logger.Debug("Verifying output directory: %s", hostfilesDir)
	if _, err := os.Stat(hostfilesDir); os.IsNotExist(err) {
		logger.Error("Output directory does not exist: %s", hostfilesDir)
		return fmt.Errorf("output directory does not exist: %s", hostfilesDir)
	}

	// Initialize a map to hold all unique IPs
	ipSet := make(map[string]struct{})

	// Helper function to read IPs from a file
	readIPs := func(filePath string) error {
		logger.Info("Reading IPs from file: %s", filePath)
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error("Failed to open file %s for categorization: %v", filePath, err)
			return err
		}
		defer file.Close()

		count := 0
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			ip := ExtractIPFromLine(line)
			if ip != "" {
				ipSet[ip] = struct{}{}
				count++
			} else {
				logger.Debug("Skipping non-IP line: %s", line)
			}
		}
		if err := scanner.Err(); err != nil {
			logger.Error("Error reading file %s: %v", filePath, err)
			return err
		}
		logger.Info("Found %d valid IPs in %s", count, filePath)
		return nil
	}

	// Read IPs from scan files with error handling
	logger.Info("Reading IPs from scan files...")
	for _, file := range []string{arpScanFile, pingScanFile, windowsDiscoveryFile} {
		if err := readIPs(file); err != nil {
			return fmt.Errorf("failed to read IPs from %s: %v", file, err)
		}
	}

	logger.Info("Total unique IPs found: %d", len(ipSet))

	// Parse ARP scan results for MAC vendor information
	logger.Info("Parsing ARP scan results for MAC vendor information")
	macInfo, err := parseARPScanResults(arpScanFile)
	if err != nil {
		logger.Error("Failed to parse ARP scan results: %v", err)
		return err
	}
	logger.Info("Found MAC vendor information for %d hosts", len(macInfo))

	// Parse Nmap XML and enrich with MAC vendor information
	logger.Info("Parsing Nmap XML file: %s", nmapXMLPath)
	nmapData, err := parseNmapXML(nmapXMLPath)
	if err != nil {
		logger.Error("Failed to parse Nmap XML: %v", err)
		return err
	}
	logger.Info("Successfully parsed Nmap data for %d hosts", len(nmapData))

	// Enrich Nmap data with MAC vendor information
	for ip, info := range nmapData {
		if macData, exists := macInfo[ip]; exists {
			info.MAC = macData.MAC
			info.MACVendor = macData.Vendor
			nmapData[ip] = info
			logger.Debug("Enriched host data for IP %s with MAC vendor: %s", ip, macData.Vendor)
		}
	}

	// Prepare category maps
	categories := map[Category][]string{
		CategoryWindowsServer:       {},
		CategoryWindowsClient:       {},
		CategoryWindowsUnknown:      {},
		CategoryLinux:               {},
		CategoryPrinter:             {},
		CategoryUPS:                 {},
		CategoryNAS:                 {},
		CategoryFirewall:            {},
		CategoryRouterSwitch:        {},
		CategoryLightsOutManagement: {},
		CategoryUnknown:             {},
	}
	logger.Info("Successfully parsed Nmap data for %d hosts", len(nmapData))

	// Categorize each IP
	logger.Info("Starting IP categorization")
	categoryCounts := make(map[Category]int)
	for ip := range ipSet {
		host, exists := nmapData[ip]
		if !exists {
			host = HostInfo{
				IP: ip,
			}
		}
		category := categorizeIP(host)
		categories[category] = append(categories[category], ip)
		categoryCounts[category]++
		logger.Debug("Categorized IP %s as %s", ip, category)
	}

	// Log category counts
	for category, count := range categoryCounts {
		logger.Info("Category %s has %d IPs", category, count)
	}

	// Define file paths for each category
	categoryFiles := map[Category]string{
		CategoryWindowsServer:       "Windows_Server.txt",
		CategoryWindowsClient:       "Windows_Client.txt",
		CategoryWindowsUnknown:      "Windows_Unknown.txt",
		CategoryLinux:               "Linux.txt",
		CategoryPrinter:             "Printers.txt",
		CategoryUPS:                 "UPS.txt",
		CategoryNAS:                 "NAS.txt",
		CategoryFirewall:            "Firewalls.txt",
		CategoryRouterSwitch:        "Routers_Switches.txt",
		CategoryLightsOutManagement: "Lights_Out_Management.txt",
		CategoryUnknown:             "Unknown.txt",
	}

	// Write results to files
	logger.Info("Writing categorized IPs to files in: %s", hostfilesDir)
	for category, ips := range categories {
		if len(ips) == 0 {
			continue
		}

		filePath := filepath.Join(hostfilesDir, categoryFiles[category])
		logger.Debug("Writing %d IPs to %s", len(ips), filePath)

		file, err := os.Create(filePath)
		if err != nil {
			logger.Error("Failed to create category file %s: %v", filePath, err)
			continue
		}

		writer := bufio.NewWriter(file)
		for _, ip := range ips {
			fmt.Fprintln(writer, ip)
		}
		writer.Flush()
		file.Close() // Close each file after writing

		logger.Info("Successfully wrote %d IPs to %s", len(ips), filePath)
	}

	logger.Info("Host categorization completed successfully")
	return nil
}

// parseNmapXML parses the Nmap XML file and returns a map of IP to HostInfo
func parseNmapXML(xmlPath string) (map[string]HostInfo, error) {
	file, err := os.Open(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Nmap XML file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Nmap XML file: %v", err)
	}

	var nmapRun parsers.NmapRun
	if err := xml.Unmarshal(data, &nmapRun); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Nmap XML: %v", err)
	}

	hostMap := make(map[string]HostInfo)
	for _, host := range nmapRun.Hosts {
		if host.Status.State != "up" {
			continue
		}

		// Build host info with enhanced OS and service details
		hostInfo := HostInfo{
			IP:        host.Address.Addr,
			HostNames: make([]string, 0),
			Ports:     make([]Port, 0),
		}

		// Parse OS information
		if len(host.OSMatches) > 0 {
			bestMatch := host.OSMatches[0] // Highest accuracy match
			hostInfo.OS.Vendor = bestMatch.OSClasses[0].Vendor
			hostInfo.OS.Family = bestMatch.OSClasses[0].OSFamily
			hostInfo.OS.Gen = bestMatch.OSClasses[0].OSGen
			hostInfo.OS.Detail = bestMatch.Name
			hostInfo.OS.Accuracy = bestMatch.Accuracy
		}

		// Parse ports and services
		for _, port := range host.Ports {
			hostInfo.Ports = append(hostInfo.Ports, Port{
				Protocol: port.Protocol,
				PortID:   port.PortId,
				State:    port.State.State,
				Service: Service{
					Name:      port.Service.Name,
					Product:   port.Service.Product,
					Version:   port.Service.Version,
					ExtraInfo: port.Service.Extra,
				},
			})
		}

		hostMap[host.Address.Addr] = hostInfo
	}

	return hostMap, nil
}

// isUPSDevice checks if a host is a UPS device based on common management ports and vendor signatures
func isUPSDevice(host HostInfo) bool {
	hasWebInterface := false
	hasUPSVendor := false

	// Check OS/Vendor information from Nmap
	vendorLower := strings.ToLower(host.OS.Vendor)
	for _, vendor := range upsVendors {
		if strings.Contains(vendorLower, vendor) {
			hasUPSVendor = true
			break
		}
	}

	// Check ports and services if vendor not found in OS info
	for _, port := range host.Ports {
		// Check for web management interface
		if (port.PortID == 80 || port.PortID == 443) && port.State == "open" {
			hasWebInterface = true
		}

		// If vendor not yet found, check service banners
		if !hasUPSVendor {
			serviceLower := strings.ToLower(port.Service.Product)
			for _, vendor := range upsVendors {
				if strings.Contains(serviceLower, vendor) {
					hasUPSVendor = true
					break
				}
			}
		}

		// Check SNMP info if vendor still not found
		if !hasUPSVendor && port.PortID == 161 && port.State == "open" {
			snmpInfo := strings.ToLower(port.Service.ExtraInfo)
			for _, vendor := range upsVendors {
				if strings.Contains(snmpInfo, vendor) {
					hasUPSVendor = true
					break
				}
			}
		}
	}

	return hasWebInterface && hasUPSVendor
}

// isFirewallDevice checks if a host is a firewall device based on common management ports and vendor signatures
func isFirewallDevice(host HostInfo) bool {
	// Check OS/Vendor information from Nmap
	vendorLower := strings.ToLower(host.OS.Vendor)
	for _, vendor := range firewallVendors {
		if strings.Contains(vendorLower, vendor) {
			return true
		}
	}

	hasWebInterface := false
	hasFirewallVendor := false
	hasFirewallPorts := false

	for _, port := range host.Ports {
		// Check for management interface (HTTPS usually preferred for firewalls)
		if (port.PortID == 443 || port.PortID == 80) && port.State == "open" {
			hasWebInterface = true
		}

		// Check for common firewall management ports
		if (port.PortID == 22 || // SSH
			port.PortID == 161 || // SNMP
			port.PortID == 162 || // SNMP Trap
			port.PortID == 500 || // IPsec IKE
			port.PortID == 4443 || // Alternative HTTPS mgmt
			port.PortID == 8443) && // Alternative HTTPS mgmt
			port.State == "open" {
			hasFirewallPorts = true
		}

		// Check service banners
		serviceLower := strings.ToLower(port.Service.Product)
		for _, vendor := range firewallVendors {
			if strings.Contains(serviceLower, vendor) {
				hasFirewallVendor = true
				break
			}
		}

		// Additional check for SSL/TLS certificate subjects
		if (port.PortID == 443 || port.PortID == 8443) &&
			strings.Contains(strings.ToLower(port.Service.ExtraInfo), "firewall") {
			hasFirewallVendor = true
		}
	}

	// A firewall typically needs:
	// 1. A web interface
	// 2. Either a known vendor signature OR typical firewall ports
	return hasWebInterface && (hasFirewallVendor || hasFirewallPorts)
}

// isPrinterDevice checks if a host is a printer based on common ports and vendor signatures
func isPrinterDevice(host HostInfo) bool {
	// Check MAC vendor first
	if host.MACVendor != "" {
		for vendor := range macVendors {
			if strings.Contains(host.MACVendor, vendor) && macVendors[vendor] == CategoryPrinter {
				logger.Debug("Detected printer by MAC vendor: %s (%s)", host.IP, host.MACVendor)
				return true
			}
		}
	}

	// Continue with existing printer detection logic
	hasPrinterPort := false
	hasPrinterVendor := false

	// Check OS/Vendor information from Nmap
	vendorLower := strings.ToLower(host.OS.Vendor)
	for _, vendor := range printerVendors {
		if strings.Contains(vendorLower, vendor) {
			hasPrinterVendor = true
			break
		}
	}

	// Check ports and services
	for _, port := range host.Ports {
		// Check for common printer ports
		if (port.PortID == 9100 || // RAW/JetDirect
			port.PortID == 515 || // LPR/LPD
			port.PortID == 631) && // IPP/CUPS
			port.State == "open" {
			hasPrinterPort = true
		}

		// If vendor not yet found, check service banners
		if !hasPrinterVendor {
			serviceLower := strings.ToLower(port.Service.Product)
			for _, vendor := range printerVendors {
				if strings.Contains(serviceLower, vendor) {
					hasPrinterVendor = true
					break
				}
			}
		}

		// Check SNMP info if vendor still not found
		if !hasPrinterVendor && port.PortID == 161 && port.State == "open" {
			snmpInfo := strings.ToLower(port.Service.ExtraInfo)
			for _, vendor := range printerVendors {
				if strings.Contains(snmpInfo, vendor) {
					hasPrinterVendor = true
					break
				}
			}
		}
	}

	return hasPrinterPort && hasPrinterVendor
}

// Update the categorizeIP function to use the new UPS detection
func categorizeIP(host HostInfo) Category {
	// Check MAC vendor first if available
	if host.MACVendor != "" {
		logger.Debug("Checking MAC vendor for %s: %s", host.IP, host.MACVendor)

		// Check for printer by MAC vendor
		if isPrinterDevice(host) {
			logger.Debug("Detected printer device by MAC vendor: %s", host.IP)
			return CategoryPrinter
		}

		// Check for UPS by MAC vendor
		if isUPSDevice(host) {
			logger.Debug("Detected UPS device by MAC vendor: %s", host.IP)
			return CategoryUPS
		}

		// Check for firewall by MAC vendor
		if isFirewallDevice(host) {
			logger.Debug("Detected firewall device by MAC vendor: %s", host.IP)
			return CategoryFirewall
		}

		// Check for network device by MAC vendor
		if isNetwork, vendor := isNetworkDevice(host); isNetwork {
			logger.Debug("Detected network device by MAC vendor: %s (vendor: %s)", host.IP, vendor)
			networkDevices[vendor] = append(networkDevices[vendor], host.IP)
			return CategoryRouterSwitch
		}
	}

	// Continue with existing OS-based categorization
	if host.OS.Accuracy >= 80 {
		logger.Debug("High confidence OS match for %s: %s (accuracy: %d)", host.IP, host.OS.Family, host.OS.Accuracy)
		osFamily := strings.ToLower(host.OS.Family)
		osDetail := strings.ToLower(host.OS.Detail)

		switch {
		case strings.Contains(osFamily, "windows"):
			if strings.Contains(osDetail, "server") {
				return CategoryWindowsServer
			} else if strings.Contains(osDetail, "windows 10") ||
				strings.Contains(osDetail, "windows 11") {
				return CategoryWindowsClient
			}
			return CategoryWindowsUnknown

		case strings.Contains(osFamily, "linux"):
			return CategoryLinux
		}
	}

	// Check for Windows systems based on open ports and services
	hasPort445 := false
	hasNetBIOS := false
	for _, port := range host.Ports {
		switch port.PortID {
		case 445:
			hasPort445 = true
		case 135, 137, 138, 139:
			hasNetBIOS = true
		}
	}
	if hasPort445 && hasNetBIOS {
		return CategoryWindowsUnknown
	}

	// Check for Linux/Unix systems based on SSH and other services
	hasSSH := false
	hasLinuxServices := false
	for _, port := range host.Ports {
		if port.Service.Name == "ssh" {
			hasSSH = true
		}
		// Look for common Linux/Unix services
		if port.Service.Name == "nfs" ||
			port.Service.Name == "mountd" ||
			port.Service.Name == "rpcbind" {
			hasLinuxServices = true
		}
	}
	if hasSSH && hasLinuxServices {
		return CategoryLinux
	}

	// Check for UPS devices first using the enhanced detection
	if isUPSDevice(host) {
		logger.Debug("Detected UPS device: %s", host.IP)
		return CategoryUPS
	}

	if isFirewallDevice(host) {
		logger.Debug("Detected firewall device: %s", host.IP)
		return CategoryFirewall
	}

	if isNetwork, vendor := isNetworkDevice(host); isNetwork {
		logger.Debug("Detected network device: %s (vendor: %s)", host.IP, vendor)
		networkDevices[vendor] = append(networkDevices[vendor], host.IP)
		return CategoryRouterSwitch
	}

	if isPrinterDevice(host) {
		logger.Debug("Detected printer device: %s", host.IP)
		return CategoryPrinter
	}

	logger.Debug("Unable to categorize IP %s, marking as unknown", host.IP)
	return CategoryUnknown
}
