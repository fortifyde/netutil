package functions

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/fortifyde/netutil/internal/functions/scanners"
	"github.com/fortifyde/netutil/internal/functions/utils"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// StartDiscoveryScan initiates the Discovery Scan process.
func StartDiscoveryScan(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Starting Discovery Scan")

	var criticalError error

	// Function to perform scans
	performScans := func(ipRange, hostfilesDir, nmapDir string) {
		// 5a. ARP Scan
		logger.Info("Starting ARP Scan on %s", ipRange)
		arpScanFile := filepath.Join(hostfilesDir, "arp_scan.txt")
		if err := scanners.PerformARPscan(ipRange, arpScanFile); err != nil {
			logger.Error("ARP Scan failed: %v", err)
			uiutil.ShowError(app, pages, fmt.Sprintf("ARP Scan failed: %v", err), mainView, nil)
			// Non-critical: Continue scanning
		}

		// 5b. Asynchronous Ping Scan using fping
		logger.Info("Starting Ping Scan on %s", ipRange)
		pingScanFile := filepath.Join(hostfilesDir, "ping_scan.txt")
		if err := scanners.PerformPingScan(ipRange, pingScanFile); err != nil {
			logger.Error("Ping Scan failed: %v", err)
			uiutil.ShowError(app, pages, fmt.Sprintf("Ping Scan failed: %v", err), mainView, nil)
			// Non-critical: Continue scanning
		}

		// 5c. DNS Reverse Lookup with dig
		logger.Info("Starting DNS Reverse Lookup on %s", pingScanFile)
		dnsLookupFile := filepath.Join(hostfilesDir, "dns_reverse_lookup.txt")
		if err := scanners.PerformDNSReverseLookup(pingScanFile, dnsLookupFile); err != nil {
			logger.Error("DNS Reverse Lookup failed: %v", err)
			uiutil.ShowError(app, pages, fmt.Sprintf("DNS Reverse Lookup failed: %v", err), mainView, nil)
			// Non-critical: Continue scanning
		}

		// 5d. Windows OS Quick Discovery
		logger.Info("Starting Windows OS Quick Discovery")
		windowsDiscoveryFile := filepath.Join(hostfilesDir, "windows_os_discovery.txt")
		if err := scanners.PerformWindowsOSDiscovery(ipRange, windowsDiscoveryFile); err != nil {
			logger.Error("Windows OS Discovery failed: %v", err)
			uiutil.ShowError(app, pages, fmt.Sprintf("Windows OS Discovery failed: %v", err), mainView, nil)
			// Non-critical: Continue scanning
		}

		// Step 6: Create Hostfile with all found IP addresses
		logger.Info("Creating hostfile")
		hostfilePath := filepath.Join(hostfilesDir, "hostfile.txt")
		if err := utils.CreateHostfile([]string{arpScanFile, pingScanFile, dnsLookupFile, windowsDiscoveryFile}, hostfilePath); err != nil {
			logger.Error("Failed to create hostfile: %v", err)
			criticalError = fmt.Errorf("failed to create hostfile: %v", err)
			return
		}

		// Step 7: Start Nmap Discovery Scan
		logger.Info("Starting Nmap Discovery Scan")
		nmapOutputDir := nmapDir
		if err := utils.EnsureDir(nmapOutputDir); err != nil {
			logger.Error("Failed to ensure Nmap directory exists: %v", err)
			criticalError = fmt.Errorf("failed to ensure Nmap directory exists: %v", err)
			return
		}

		nmapScanFile := filepath.Join(nmapOutputDir, "nmap_discovery_scan.xml")
		if err := scanners.PerformNmapScan(hostfilePath, nmapScanFile); err != nil {
			logger.Error("Nmap Discovery Scan failed: %v", err)
			uiutil.ShowError(app, pages, fmt.Sprintf("Nmap Discovery Scan failed: %v", err), mainView, nil)
			// Non-critical: Continue
		}

		// Step 8: Notify the user of completion
		uiutil.ShowMessage(app, pages, "Discovery Scan completed successfully.", mainView)
		logger.Info("Discovery Scan completed successfully")
	}

	// Function to proceed with scan or handle critical errors
	var proceedWithScan func(ipRange, selectedInterface, vlanID string)
	proceedWithScan = func(ipRange, selectedInterface, vlanID string) {
		// Step 3: Ask user to enter a name for the directory
		uiutil.PromptInput(app, pages, "Discovery Scan", "Enter directory name for hostfiles:", mainView, func(dirName string, err error) {
			if err != nil {
				// User canceled the prompt
				logger.Info("User canceled the Discovery Scan.")
				return
			}

			dirName = strings.TrimSpace(dirName)
			if dirName == "" {
				uiutil.ShowError(app, pages, "Directory name cannot be empty. Please enter a valid name.", mainView, nil)
				proceedWithScan(ipRange, selectedInterface, vlanID)
				return
			}

			logger.Info("User provided directory name: %s", dirName)

			// Step 4: Create necessary directories
			workingDir := utils.GetWorkingDirectory()
			hostfilesDir := filepath.Join(workingDir, "#Hostfiles", dirName)
			nmapDir := filepath.Join(workingDir, "Nmap", dirName)

			dirs := []string{hostfilesDir, nmapDir}
			for _, dir := range dirs {
				if err := utils.EnsureDir(dir); err != nil {
					logger.Error("Failed to create directory %s: %v", dir, err)
					criticalError = fmt.Errorf("failed to create directory %s: %v", dir, err)
					return
				}
			}
			logger.Info("Created directories: %s, %s", hostfilesDir, nmapDir)

			// Step 5: Perform scans
			performScans(ipRange, hostfilesDir, nmapDir)
		})
	}

	// Function to prompt for Interface
	var promptInterface func(ipRange string)
	promptInterface = func(ipRange string) {
		uiutil.PromptInput(app, pages, "Discovery Scan", "Enter interface (e.g., eth0 or eth0.10):", mainView, func(input string, err error) {
			if err != nil {
				// User canceled the prompt
				logger.Info("User canceled the Discovery Scan.")
				return
			}

			input = strings.TrimSpace(input)
			if input == "" {
				uiutil.ShowError(app, pages, "Interface cannot be empty. Please enter a valid interface.", mainView, nil)
				promptInterface(ipRange)
				return
			}

			parts := strings.Split(input, ".")
			var interfaceName, vlanID string
			if len(parts) == 2 {
				interfaceName = parts[0]
				vlanID = parts[1]
			} else {
				interfaceName = input
			}

			logger.Info("User provided interface: %s, VLAN ID: %s", interfaceName, vlanID)
			proceedWithScan(ipRange, interfaceName, vlanID)
		})
	}

	// Function to prompt for IP range
	var promptIPRange func()
	promptIPRange = func() {
		uiutil.PromptInput(app, pages, "Discovery Scan", "Enter IP range to scan (e.g., 192.168.1.0/24):", mainView, func(ipRange string, err error) {
			if err != nil {
				// User canceled the prompt
				logger.Info("User canceled the Discovery Scan.")
				return
			}
			ipRange = strings.TrimSpace(ipRange)
			if ipRange == "" {
				uiutil.ShowError(app, pages, "IP range cannot be empty. Please enter a valid IP range.", mainView, nil)
				promptIPRange()
				return
			}
			logger.Info("User provided IP range: %s", ipRange)

			// Validate IP range
			if _, _, err := net.ParseCIDR(ipRange); err != nil {
				uiutil.ShowError(app, pages, "Invalid IP range format. Please try again.", mainView, nil)
				promptIPRange()
				return
			}

			// Proceed to detect interface
			detectedInterface, vlanID, err := detectInterfaceForIPRange(ipRange)
			if err != nil {
				logger.Warning("Failed to detect interface automatically: %v", err)
			}

			if detectedInterface != "" {
				// Ask user to confirm the detected interface
				uiutil.PromptConfirmation(app, pages, "Discovery Scan", fmt.Sprintf("Detected interface for IP range %s: %s. Do you want to use this interface?", ipRange, detectedInterface), func(confirm bool, err error) {
					if err != nil || !confirm {
						// User declined, prompt to enter their own choice
						promptInterface(ipRange)
						return
					}
					// User confirmed the detected interface
					selectedInterface := detectedInterface
					proceedWithScan(ipRange, selectedInterface, vlanID)
				}, mainView)
			} else {
				// No interface detected, prompt user to enter their own choice
				promptInterface(ipRange)
			}
		})
	}

	// Start the prompting process
	promptIPRange()

	// Check for critical errors
	if criticalError != nil {
		return criticalError
	}

	return nil
}

// detectInterfaceForIPRange attempts to automatically detect the appropriate interface or VLAN for the given IP range.
func detectInterfaceForIPRange(ipRange string) (interfaceName string, vlanID string, err error) {
	// Parse the IP range to get the network
	_, network, err := net.ParseCIDR(ipRange)
	if err != nil {
		logger.Warning("Invalid IP range format: %v", err)
		return "", "", fmt.Errorf("invalid IP range format")
	}

	interfaces, err := utils.GetEthernetInterfaces()
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
