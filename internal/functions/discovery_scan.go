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

	// Step 1: Ask the user for an IP range to scan
	var ipRange string
	uiutil.PromptInput(app, pages, "Discovery Scan", "Enter IP range to scan (e.g., 192.168.1.0/24):", mainView, func(input string, err error) {
		if err != nil {
			logger.Error("User canceled IP range input: %v", err)
			ipRange = ""
			return
		}
		ipRange = input
		logger.Info("User provided IP range: %s", ipRange)
	})
	if ipRange == "" {
		return fmt.Errorf("IP range input canceled or empty")
	}

	// Step 2: Automatically attribute to a fitting interface or VLAN subinterface
	detectedInterface, vlanID, err := detectInterfaceForIPRange(ipRange)
	if err != nil {
		logger.Warning("Failed to detect interface automatically: %v", err)
	}

	var selectedInterface string
	if detectedInterface != "" {
		// Ask user to confirm the detected interface
		confirm, err := uiutil.PromptConfirmation(app, pages, "Discovery Scan", fmt.Sprintf("Detected interface for IP range %s: %s", ipRange, detectedInterface))
		if err != nil || !confirm {
			// User declined, prompt to enter their own choice
			selectedInterface, vlanID, err = promptUserForInterface(app, pages, "Discovery Scan", mainView)
			if err != nil {
				logger.Error("User failed to provide interface: %v", err)
				return fmt.Errorf("interface selection failed: %v", err)
			}
		} else {
			selectedInterface = detectedInterface
		}
	} else {
		// No interface detected, prompt user to enter their own choice
		selectedInterface, vlanID, err = promptUserForInterface(app, pages, "Discovery Scan", mainView)
		if err != nil {
			logger.Error("User failed to provide interface: %v", err)
			return fmt.Errorf("interface selection failed: %v", err)
		}
	}

	logger.Info("Selected interface: %s", selectedInterface)

	// Step 3: Ask user to enter a name for the directory
	var dirName string
	uiutil.PromptInput(app, pages, "Discovery Scan", "Enter directory name for hostfiles:", mainView, func(input string, err error) {
		if err != nil {
			logger.Error("User canceled directory name input: %v", err)
			dirName = ""
			return
		}
		if input == "" {
			logger.Warning("Empty directory name provided")
			dirName = ""
			return
		}
		dirName = input
		logger.Info("User provided directory name: %s", dirName)
	}, vlanID)
	if dirName == "" {
		return fmt.Errorf("directory name input canceled or empty")
	}

	// Step 4: Create necessary directories
	workingDir := utils.GetWorkingDirectory()
	hostfilesDir := filepath.Join(workingDir, "#Hostfiles", dirName)
	nmapDir := filepath.Join(workingDir, "Nmap", dirName)

	dirs := []string{hostfilesDir, nmapDir}
	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			logger.Error("Failed to create directory %s: %v", dir, err)
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	logger.Info("Created directories: %s, %s", hostfilesDir, nmapDir)

	// Step 5: Perform scans
	// 5a. ARP Scan
	logger.Info("Starting ARP Scan on %s", ipRange)
	arpScanFile := filepath.Join(hostfilesDir, "arp_scan.txt")
	if err := scanners.PerformARPscan(ipRange, arpScanFile); err != nil {
		logger.Error("ARP Scan failed: %v", err)
		uiutil.ShowError(app, pages, fmt.Sprintf("ARP Scan failed: %v", err), mainView, nil)
	}

	// 5b. Asynchronous Ping Scan using fping
	logger.Info("Starting Ping Scan on %s", ipRange)
	pingScanFile := filepath.Join(hostfilesDir, "ping_scan.txt")
	if err := scanners.PerformPingScan(ipRange, pingScanFile); err != nil {
		logger.Error("Ping Scan failed: %v", err)
		uiutil.ShowError(app, pages, fmt.Sprintf("Ping Scan failed: %v", err), mainView, nil)
	}

	// 5c. DNS Reverse Lookup with dig
	logger.Info("Starting DNS Reverse Lookup on %s", pingScanFile)
	dnsLookupFile := filepath.Join(hostfilesDir, "dns_reverse_lookup.txt")
	if err := scanners.PerformDNSReverseLookup(pingScanFile, dnsLookupFile); err != nil {
		logger.Error("DNS Reverse Lookup failed: %v", err)
		uiutil.ShowError(app, pages, fmt.Sprintf("DNS Reverse Lookup failed: %v", err), mainView, nil)
	}

	// 5d. Windows OS Quick Discovery
	logger.Info("Starting Windows OS Quick Discovery")
	windowsDiscoveryFile := filepath.Join(hostfilesDir, "windows_os_discovery.txt")
	if err := scanners.PerformWindowsOSDiscovery(ipRange, windowsDiscoveryFile); err != nil {
		logger.Error("Windows OS Discovery failed: %v", err)
		uiutil.ShowError(app, pages, fmt.Sprintf("Windows OS Discovery failed: %v", err), mainView, nil)
	}

	// Step 6: Create Hostfile with all found IP addresses
	logger.Info("Creating hostfile")
	hostfilePath := filepath.Join(hostfilesDir, "hostfile.txt")
	if err := utils.CreateHostfile([]string{arpScanFile, pingScanFile, dnsLookupFile, windowsDiscoveryFile}, hostfilePath); err != nil {
		logger.Error("Failed to create hostfile: %v", err)
		uiutil.ShowError(app, pages, fmt.Sprintf("Failed to create hostfile: %v", err), mainView, nil)
	}

	// Step 7: Start Nmap Discovery Scan
	logger.Info("Starting Nmap Discovery Scan")
	nmapOutputDir := nmapDir
	if err := utils.EnsureDir(nmapOutputDir); err != nil {
		logger.Error("Failed to ensure Nmap directory exists: %v", err)
		return fmt.Errorf("failed to ensure Nmap directory exists: %v", err)
	}

	nmapScanFile := filepath.Join(nmapOutputDir, "nmap_discovery_scan.xml")
	if err := scanners.PerformNmapScan(hostfilePath, nmapScanFile); err != nil {
		logger.Error("Nmap Discovery Scan failed: %v", err)
		uiutil.ShowError(app, pages, fmt.Sprintf("Nmap Discovery Scan failed: %v", err), mainView, nil)
	}

	// Step 8: Notify the user of completion
	uiutil.ShowMessage(app, pages, "Discovery Scan completed successfully.", mainView)
	logger.Info("Discovery Scan completed successfully")
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

// promptUserForInterface prompts the user to enter an interface in the format "interfaceName[.vlanID]".
func promptUserForInterface(app *tview.Application, pages *tview.Pages, title string, mainView tview.Primitive) (interfaceName string, vlanID string, err error) {
	done := make(chan bool)
	var result string
	var inputErr error

	uiutil.PromptInput(app, pages, title, "Enter interface (e.g., eth0 or eth0.10):", mainView, func(input string, err error) {
		if err != nil {
			inputErr = err
			done <- true
			return
		}
		result = input
		done <- true
	})

	<-done

	if inputErr != nil {
		return "", "", inputErr
	}

	if result == "" {
		return "", "", fmt.Errorf("interface input cannot be empty")
	}

	parts := strings.Split(result, ".")
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return result, "", nil
}
