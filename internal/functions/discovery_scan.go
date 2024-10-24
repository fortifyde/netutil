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

// initiates the Discovery Scan process.
func StartDiscoveryScan(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Starting Discovery Scan")

	var criticalError error
	var promptIPRange func()
	var promptInterface func(ipRange string)
	var promptDirName func(vlanID string)
	performScans := func(ipRange, selectedInterface, vlanID, hostfilesDir, nmapDir string) {
		logger.Info("Starting ARP Scan on %s", ipRange)
		arpScanFile := filepath.Join(hostfilesDir, "arp_scan.txt")
		if err := scanners.PerformARPscan(ipRange, selectedInterface, vlanID, arpScanFile); err != nil {
			logger.Error("ARP Scan failed: %v", err)
			uiutil.ShowError(app, pages, "arpScanErrorModal",
				fmt.Sprintf("ARP Scan failed: %v", err),
				mainView,
				nil)
		}

		logger.Info("Starting Ping Scan on %s", ipRange)
		pingScanFile := filepath.Join(hostfilesDir, "ping_scan.txt")
		if err := scanners.PerformPingScan(ipRange, selectedInterface, vlanID, pingScanFile); err != nil {
			logger.Error("Ping Scan failed: %v", err)
			uiutil.ShowError(app, pages, "pingScanErrorModal",
				fmt.Sprintf("Ping Scan failed: %v", err),
				mainView,
				nil)
		}

		logger.Info("Starting DNS Reverse Lookup on %s", pingScanFile)
		dnsLookupFile := filepath.Join(hostfilesDir, "dns_reverse_lookup.txt")
		if err := scanners.PerformDNSReverseLookup(pingScanFile, dnsLookupFile); err != nil {
			logger.Error("DNS Reverse Lookup failed: %v", err)
			uiutil.ShowError(app, pages, "dnsReverseLookupErrorModal",
				fmt.Sprintf("DNS Reverse Lookup failed: %v", err),
				mainView,
				nil)
		}

		logger.Info("Starting Windows OS Quick Discovery")
		windowsDiscoveryFile := filepath.Join(hostfilesDir, "windows_os_discovery.txt")
		if err := scanners.PerformWindowsOSDiscovery(pingScanFile, windowsDiscoveryFile); err != nil {
			logger.Error("Windows OS Discovery failed: %v", err)
			uiutil.ShowError(app, pages, "windowsOSDiscoveryErrorModal",
				fmt.Sprintf("Windows OS Discovery failed: %v", err),
				mainView,
				nil)
		}

		logger.Info("Creating hostfile")
		hostfilePath := filepath.Join(hostfilesDir, "hostfile.txt")
		if err := utils.CreateHostfile([]string{arpScanFile, pingScanFile, dnsLookupFile, windowsDiscoveryFile}, hostfilePath); err != nil {
			logger.Error("Failed to create hostfile: %v", err)
			criticalError = fmt.Errorf("failed to create hostfile: %v", err)
			return
		}

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
			uiutil.ShowError(app, pages, "nmapDiscoveryScanErrorModal",
				fmt.Sprintf("Nmap Discovery Scan failed: %v", err),
				mainView,
				nil)
		}

		uiutil.ShowMessage(app, pages, "scan_complete_modal", "Discovery Scan completed successfully.", mainView)
		logger.Info("Discovery Scan completed successfully")
	}

	proceedWithScan := func(ipRange, selectedInterface, vlanID string) {
		dirNameCallback := func(dirName string, err error) {
			if err != nil {
				logger.Info("User canceled the Discovery Scan.")
				return
			}

			dirName = strings.TrimSpace(dirName)
			if dirName == "" {
				uiutil.ShowError(app, pages, "dirNameErrorModal",
					"Directory name cannot be empty. Please enter a valid name.",
					mainView,
					func() {
						promptDirName(vlanID)
					})
				return
			}

			logger.Info("User provided directory name: %s", dirName)

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

			performScans(ipRange, selectedInterface, vlanID, hostfilesDir, nmapDir)
		}

		promptDirName = func(vlanID string) {
			uiutil.PromptInput(app, pages, "dirNameInputModal", "Discovery Scan",
				"Enter directory name for hostfiles:", mainView,
				dirNameCallback, "vlan"+vlanID)
		}

		promptDirName(vlanID)
	}

	promptIPRange = func() {
		uiutil.PromptInput(app, pages, "ipRangeInputModal", "Discovery Scan",
			"Enter IP range to scan (e.g., 192.168.1.0/24):", mainView,
			func(ipRange string, err error) {
				if err != nil {
					logger.Info("User canceled the Discovery Scan.")
					return
				}
				ipRange = strings.TrimSpace(ipRange)
				if ipRange == "" {
					uiutil.ShowError(app, pages, "ipRangeEmptyErrorModal",
						"IP range cannot be empty. Please enter a valid IP range.",
						mainView,
						func() {
							promptIPRange()
						})
					return
				}
				logger.Info("User provided IP range: %s", ipRange)

				if _, _, err := net.ParseCIDR(ipRange); err != nil {
					uiutil.ShowError(app, pages, "ipRangeFormatErrorModal",
						"Invalid IP range format. Please try again.",
						mainView,
						func() {
							promptIPRange()
						})
					return
				}

				detectedInterface, vlanID, err := detectInterfaceForIPRange(ipRange)
				if err != nil {
					logger.Warning("Failed to detect interface automatically: %v", err)
				}

				if detectedInterface != "" {
					uiutil.ShowConfirm(app, pages, "confirmInterfaceModal",
						fmt.Sprintf("Detected interface for IP range %s: %s. Do you want to use this interface?", ipRange, detectedInterface),
						func(confirm bool) {
							if !confirm {
								promptInterface(ipRange)
								return
							}
							selectedInterface := detectedInterface
							proceedWithScan(ipRange, selectedInterface, vlanID)
						}, mainView)
				} else {
					promptInterface(ipRange)
				}
			})
	}

	promptInterface = func(ipRange string) {
		uiutil.PromptInput(app, pages, "interfaceInputModal", "Select Interface",
			"Please enter the interface to use:", mainView,
			func(selectedInterface string, err error) {
				if err != nil {
					logger.Info("User canceled the Discovery Scan.")
					return
				}
				selectedInterface = strings.TrimSpace(selectedInterface)
				if selectedInterface == "" {
					uiutil.ShowError(app, pages, "interfaceEmptyErrorModal",
						"Interface cannot be empty. Please enter a valid interface.",
						mainView,
						func() {
							promptInterface(ipRange)
						})
					return
				}

				logger.Info("User selected interface: %s", selectedInterface)
				proceedWithScan(ipRange, selectedInterface, "")
			})
	}

	promptIPRange()

	if criticalError != nil {
		return criticalError
	}

	return nil
}

func detectInterfaceForIPRange(ipRange string) (interfaceName string, vlanID string, err error) {
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
			default:
				continue
			}
			if network.Contains(ip) {
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
