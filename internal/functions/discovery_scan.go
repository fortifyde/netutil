package functions

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fortifyde/netutil/internal/functions/scanners"
	"github.com/fortifyde/netutil/internal/functions/utils"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// StartDiscoveryScan initiates the Discovery Scan process with sequential prompts,
// automatic interface detection, user confirmation, and real-time output display.
func StartDiscoveryScan(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Initiating Discovery Scan process")

	// Create a cancellable context for the scanning operations
	ctx, cancel := context.WithCancel(context.Background())

	// Variables to store user inputs
	var ipRange, selectedInterface, vlanID, dirName string
	var dirNameCallback func(input string, err error)
	var interfaceConfirmationCallback func(confirm bool, err error)
	var ipRangeCallback func(input string, err error)
	// Initialize WaitGroup
	var wg sync.WaitGroup
	var outputModal *uiutil.OutputModal

	// Callback after Directory Name is inputted
	dirNameCallback = func(input string, err error) {
		if err != nil {
			uiutil.ShowError(app, pages, "dirNameErrorModal",
				"Directory name cannot be empty. Please enter a valid name.",
				mainView,
				func() {
					uiutil.PromptInput(app, pages, "dirNameInputModal", "Discovery Scan",
						"Enter directory name for hostfiles:", mainView,
						dirNameCallback, "vlan"+vlanID)
				})
			return
		}

		dirName = strings.TrimSpace(input)
		if dirName == "" {
			uiutil.ShowError(app, pages, "dirNameEmptyErrorModal",
				"Directory name cannot be empty. Please enter a valid name.",
				mainView,
				func() {
					uiutil.PromptInput(app, pages, "dirNameInputModal", "Discovery Scan",
						"Enter directory name for hostfiles:", mainView,
						dirNameCallback, "vlan"+vlanID)
				})
			return
		}

		logger.Info("Directory Name provided: %s", dirName)

		// Proceed to start scanning
		wg.Add(1)
		go startScanning(app, pages, ctx, cancel, dirName, ipRange, selectedInterface, vlanID, &wg, outputModal)
	}

	// Callback after Interface and VLAN ID confirmation
	interfaceConfirmationCallback = func(confirm bool, err error) {
		if err != nil {
			// Handle cancellation
			cancel()
			return
		}
		if confirm {
			// User confirmed the detected interface and VLAN ID
			uiutil.PromptInput(app, pages, "dirNameInputModal", "Discovery Scan",
				"Enter directory name for hostfiles:", mainView,
				dirNameCallback, "vlan"+vlanID)
		} else {
			// Prompt for manual interface input
			logger.Info("User declined detected interface. Prompting for manual input.")
			uiutil.PromptInput(app, pages, "manualInterfaceInputModal", "Manual Interface Input",
				"Enter the network interface (e.g., eth0 or eth0.100):", mainView,
				func(input string, err error) {
					input = strings.TrimSpace(input)
					if err != nil {
						// Handle cancellation
						logger.Info("Manual interface input canceled by user.")
						cancel()
						return
					}

					if input == "" {
						uiutil.ShowError(app, pages, "manualInterfaceEmptyErrorModal",
							"Interface cannot be empty. Please enter a valid interface.",
							mainView,
							func() {
								// Retry manual interface input
								interfaceConfirmationCallback(false, nil)
							})
						return
					}

					// Validate the inputted interface using utility functions
					if !utils.IsValidInterface(input) {
						uiutil.ShowError(app, pages, "invalidInterfaceErrorModal",
							fmt.Sprintf("'%s' is not a valid network interface. Please enter a valid interface.", input),
							mainView,
							func() {
								// Retry manual interface input
								interfaceConfirmationCallback(false, nil)
							})
						return
					}

					// Split interface and VLAN ID if applicable
					parts := strings.Split(input, ".")
					selectedInterface = parts[0]
					if len(parts) == 2 {
						vlanID = parts[1]
					} else {
						vlanID = ""
					}

					logger.Info("User provided interface: %s, VLAN ID: %s", selectedInterface, vlanID)

					// Proceed to Directory Name prompt
					uiutil.PromptInput(app, pages, "dirNameInputModal", "Discovery Scan",
						"Enter directory name for hostfiles:", mainView,
						dirNameCallback, "vlan"+vlanID)
				}, "")
		}
	}

	// Callback after Interface Detection
	detectedInterfaceCallback := func(detectedInterface, detectedVLAN string, err error) {
		if err != nil {
			// Detection failed, prompt for manual input
			uiutil.PromptInput(app, pages, "manualInterfaceInputModal", "Manual Interface Input",
				"Enter the network interface (e.g., eth0 or eth0.100):", mainView,
				func(input string, err error) {
					input = strings.TrimSpace(input)
					if err != nil {
						// Handle cancellation
						logger.Info("Manual interface input canceled by user.")
						cancel()
						return
					}

					if input == "" {
						uiutil.ShowError(app, pages, "manualInterfaceEmptyErrorModal",
							"Interface cannot be empty. Please enter a valid interface.",
							mainView,
							func() {
								// Retry manual interface input
								interfaceConfirmationCallback(false, nil)
							})
						return
					}

					// Validate the inputted interface using utility functions
					if !utils.IsValidInterface(input) {
						uiutil.ShowError(app, pages, "invalidInterfaceErrorModal",
							fmt.Sprintf("'%s' is not a valid network interface. Please enter a valid interface.", input),
							mainView,
							func() {
								// Retry manual interface input
								interfaceConfirmationCallback(false, nil)
							})
						return
					}

					// Split interface and VLAN ID if applicable
					parts := strings.Split(input, ".")
					selectedInterface = parts[0]
					if len(parts) == 2 {
						vlanID = parts[1]
					} else {
						vlanID = ""
					}

					logger.Info("User provided interface: %s, VLAN ID: %s", selectedInterface, vlanID)

					// Proceed to Directory Name prompt
					uiutil.PromptInput(app, pages, "dirNameInputModal", "Discovery Scan",
						"Enter directory name for hostfiles:", mainView,
						dirNameCallback, "")
				}, "")
			return
		}

		// Present the detected interface and VLAN ID for user confirmation
		confirmationMessage := fmt.Sprintf("Detected Interface: %s\nDetected VLAN ID: %s\nDo you want to use these settings?",
			detectedInterface, detectedVLAN)
		selectedInterface = detectedInterface
		vlanID = detectedVLAN
		uiutil.PromptConfirmation(app, pages, "interfaceConfirmModal", "Confirm Interface",
			confirmationMessage, interfaceConfirmationCallback, mainView)
	}

	// Callback after IP Range is inputted
	ipRangeCallback = func(input string, err error) {
		if err != nil {
			// Handle cancellation
			logger.Info("IP Range input canceled by user.")
			cancel()
			return
		}

		ipRange = strings.TrimSpace(input)
		if ipRange == "" {
			uiutil.ShowError(app, pages, "ipRangeEmptyErrorModal",
				"IP Range cannot be empty. Please enter a valid IP range.",
				mainView,
				func() {
					// Retry IP Range input
					uiutil.PromptInput(app, pages, "ipRangeInputModal", "Discovery Scan",
						"Enter IP range to scan (e.g., 192.168.1.0/24):", mainView,
						ipRangeCallback, "")
				})
			return
		}

		// Attempt to detect interface and VLAN ID based on IP Range
		logger.Info("Attempting to detect interface for IP range: %s", ipRange)
		detectedInterface, detectedVLAN, err := utils.DetectInterfaceForIPRange(ipRange)
		detectedInterfaceCallback(detectedInterface, detectedVLAN, err)
	}

	// Start by prompting for IP Range
	uiutil.PromptInput(app, pages, "ipRangeInputModal", "Discovery Scan",
		"Enter IP range to scan (e.g., 192.168.1.0/24):", mainView,
		ipRangeCallback, "")

	return nil // The scanning process continues asynchronously
}

// startScanning handles the scanning operations in a separate goroutine.
func startScanning(app *tview.Application, pages *tview.Pages, ctx context.Context, cancel context.CancelFunc,
	dirName, ipRange, selectedInterface, vlanID string,
	wg *sync.WaitGroup, outputModal *uiutil.OutputModal) {

	defer wg.Done()

	// Initialize the Output Modal here
	outputModal = uiutil.ShowOutputModal(app, pages, "outputModal", "[yellow]Discovery Scan Output[-]",
		func() {
			// Handle cancellation
			logger.Info("User requested cancellation")
			cancel()
			outputModal.AppendText("[yellow]Scan canceled by user.[-]\n")
		})

	// Define the working directory paths
	workingDir := utils.GetWorkingDirectory()
	hostfilesDir := filepath.Join(workingDir, "#Hostfiles", dirName)
	scanDir := filepath.Join(hostfilesDir, "scans")

	dirs := []string{hostfilesDir, scanDir}
	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			outputModal.AppendText(fmt.Sprintf("[red]Failed to create directory %s: %v[-]\n", dir, err))
			logger.Error("Failed to create directory %s: %v", dir, err)
			cancel()
			return
		}
	}
	outputModal.AppendText(fmt.Sprintf("[blue]Created directories: %s, %s[-]\n", hostfilesDir, scanDir))

	// Perform ARP Scan
	outputModal.AppendText("[blue]Starting ARP Scan[-]\n")
	arpScanFile := filepath.Join(scanDir, "arp_scan.txt")
	if err := scanners.PerformARPscan(ctx, ipRange, selectedInterface, vlanID, arpScanFile, func(format string, a ...interface{}) {
		outputModal.AppendText(fmt.Sprintf(format, a...))
	}); err != nil {
		outputModal.AppendText(fmt.Sprintf("[red]ARP Scan failed: %v[-]\n", err))
		logger.Error("ARP Scan failed: %v", err)
		cancel()
		return
	}

	// Perform Ping Scan
	outputModal.AppendText("[blue]Starting Ping Scan[-]\n")
	pingScanFile := filepath.Join(scanDir, "ping_scan.txt")
	if err := scanners.PerformPingScan(ctx, ipRange, selectedInterface, vlanID, pingScanFile, func(format string, a ...interface{}) {
		outputModal.AppendText(fmt.Sprintf(format, a...))
	}); err != nil {
		outputModal.AppendText(fmt.Sprintf("[red]Ping Scan failed: %v[-]\n", err))
		logger.Error("Ping Scan failed: %v", err)
	}

	// Perform DNS Reverse Lookup
	outputModal.AppendText("[blue]Starting DNS Reverse Lookup[-]\n")
	dnsLookupFile := filepath.Join(scanDir, "dns_reverse_lookup.txt")
	if err := scanners.PerformDNSReverseLookup(ctx, pingScanFile, dnsLookupFile, func(format string, a ...interface{}) {
		outputModal.AppendText(fmt.Sprintf(format, a...))
	}); err != nil {
		outputModal.AppendText(fmt.Sprintf("[red]DNS Reverse Lookup failed: %v[-]\n", err))
		logger.Error("DNS Reverse Lookup failed: %v", err)
		cancel()
		return
	}

	// Perform Windows OS Discovery
	outputModal.AppendText("[blue]Starting Windows OS Discovery[-]\n")
	windowsDiscoveryFile := filepath.Join(scanDir, "windows_os_discovery.txt")
	if err := scanners.PerformWindowsOSDiscovery(ctx, pingScanFile, selectedInterface, vlanID, windowsDiscoveryFile, func(format string, a ...interface{}) {
		outputModal.AppendText(fmt.Sprintf(format, a...))
	}); err != nil {
		outputModal.AppendText(fmt.Sprintf("[red]Windows OS Discovery failed: %v[-]\n", err))
		logger.Error("Windows OS Discovery failed: %v", err)
		cancel()
		return
	}

	// Create Hostfile
	outputModal.AppendText("[blue]Creating hostfile[-]\n")
	hostfilePath := filepath.Join(scanDir, "hostfile.txt")
	if err := utils.CreateHostfile([]string{arpScanFile, pingScanFile, dnsLookupFile, windowsDiscoveryFile}, hostfilePath); err != nil {
		outputModal.AppendText(fmt.Sprintf("[red]Failed to create hostfile: %v[-]\n", err))
		logger.Error("Failed to create hostfile: %v", err)
		cancel()
		return
	}
	outputModal.AppendText(fmt.Sprintf("[blue]Hostfile created at %s[-]\n", hostfilePath))

	// Perform Nmap Discovery Scan
	outputModal.AppendText("[blue]Starting Nmap Discovery Scan[-]\n")
	nmapScanFile := filepath.Join(scanDir, "nmap_discovery_scan.xml")
	if err := scanners.PerformNmapScan(ctx, hostfilePath, selectedInterface, vlanID, nmapScanFile, func(format string, a ...interface{}) {
		outputModal.AppendText(fmt.Sprintf(format, a...))
	}); err != nil {
		outputModal.AppendText(fmt.Sprintf("[red]Nmap Discovery Scan failed: %v[-]\n", err))
		logger.Error("Nmap Discovery Scan failed: %v", err)
		cancel()
		return
	}

	// All scans completed successfully
	outputModal.AppendText("[green]All scans completed successfully.[-]\n")
	outputModal.CloseOutputModal("Discovery Scan Completed Successfully.")
}
