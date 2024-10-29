package functions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
			cancel()
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
		go startScanning(app, pages, ctx, cancel, dirName, ipRange, selectedInterface, vlanID, &wg, &outputModal, mainView)
	}

	// Callback after Interface Confirmation
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
	wg *sync.WaitGroup, outputModal **uiutil.OutputModal, mainView tview.Primitive) {

	defer wg.Done()

	// Initialize the Output Modal here
	*outputModal = uiutil.ShowOutputModal(app, pages, "outputModal", "[yellow]Discovery Scan Output[-]",
		func() {
			// Handle cancellation (Ctrl+C or 'q')
			logger.Info("User requested cancellation")
			cancel() // Signal cancellation to all goroutines
		}, mainView) // mainView is our toolbox

	// Create a channel to coordinate cleanup
	cleanupDone := make(chan struct{})

	// Start a goroutine to handle context cancellation
	go func() {
		select {
		case <-ctx.Done():
			// Context was cancelled (either by user or error)
			(*outputModal).AppendText("[yellow]Scan canceled, cleaning up...[-]\n")
			time.Sleep(2 * time.Second)
			(*outputModal).SetScanning(false)
			(*outputModal).CloseOutputModal("Scan Canceled")
		case <-cleanupDone:
			// Normal completion
			(*outputModal).SetScanning(false)
			logger.Info("All scans completed successfully")
			(*outputModal).AppendText("[green]All scans completed successfully.[-]\n")
			//(*outputModal).CloseOutputModal("Discovery Scan Completed Successfully")
		}
	}()

	// Define the working directory paths
	workingDir := utils.GetWorkingDirectory()

	// Check for existing hostfiles directory
	entries, _ := os.ReadDir(workingDir)
	hostfilesDirName := "Hostfiles"
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "hostfiles") {
			hostfilesDirName = entry.Name()
			break
		}
	}

	hostfilesDir := filepath.Join(workingDir, hostfilesDirName, dirName)
	scanDir := filepath.Join(hostfilesDir, "scans")

	// Create directories
	dirs := []string{hostfilesDir, scanDir}
	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			app.QueueUpdateDraw(func() {
				(*outputModal).AppendText(fmt.Sprintf("[red]Failed to create directory %s: %v[-]\n", dir, err))
				logger.Error("Failed to create directory %s: %v", dir, err)
				cancel()
			})
			return
		}
	}
	(*outputModal).AppendText(fmt.Sprintf("[blue]Created directories: %s, %s[-]\n", hostfilesDir, scanDir))

	// Define all scan steps
	scanSteps := []struct {
		name     string
		execute  func() error
		critical bool // If true, failure stops subsequent scans
	}{
		{
			name: "ARP Scan",
			execute: func() error {
				return scanners.PerformARPscan(ctx, ipRange, selectedInterface, vlanID,
					filepath.Join(scanDir, "arp_scan.txt"),
					func(format string, a ...interface{}) {
						(*outputModal).AppendText(fmt.Sprintf(format, a...))
					})
			},
			critical: true,
		},
		{
			name: "Ping Scan",
			execute: func() error {
				return scanners.PerformPingScan(ctx, ipRange, selectedInterface, vlanID,
					filepath.Join(scanDir, "ping_scan.txt"),
					func(format string, a ...interface{}) {
						(*outputModal).AppendText(fmt.Sprintf(format, a...))
					})
			},
			critical: false,
		},
		{
			name: "DNS Reverse Lookup",
			execute: func() error {
				return scanners.PerformDNSReverseLookup(ctx, filepath.Join(scanDir, "ping_scan.txt"),
					filepath.Join(scanDir, "dns_reverse_lookup.txt"),
					func(format string, a ...interface{}) {
						(*outputModal).AppendText(fmt.Sprintf(format, a...))
					})
			},
			critical: false,
		},
		{
			name: "Windows OS Discovery",
			execute: func() error {
				return scanners.PerformWindowsOSDiscovery(ctx, filepath.Join(scanDir, "ping_scan.txt"),
					selectedInterface, vlanID,
					filepath.Join(scanDir, "windows_os_discovery.txt"),
					func(format string, a ...interface{}) {
						(*outputModal).AppendText(fmt.Sprintf(format, a...))
					})
			},
			critical: false,
		},
		{
			name: "Create Hostfile",
			execute: func() error {
				return utils.CreateHostfile([]string{filepath.Join(scanDir, "ping_scan.txt")},
					filepath.Join(hostfilesDir, "hosts_found_up.txt"))
			},
			critical: true,
		},
		{
			name: "Nmap Discovery Scan",
			execute: func() error {
				return scanners.PerformNmapScan(ctx, filepath.Join(hostfilesDir, "hosts_found_up.txt"),
					selectedInterface, vlanID,
					filepath.Join(scanDir, "nmap_discovery_scan.xml"),
					func(format string, a ...interface{}) {
						(*outputModal).AppendText(fmt.Sprintf(format, a...))
					})
			},
			critical: true,
		},
		{
			name: "Categorize Scan Results",
			execute: func() error {
				return utils.PerformCategorization(
					hostfilesDir,
					filepath.Join(scanDir, "arp_scan.txt"),
					filepath.Join(scanDir, "ping_scan.txt"),
					filepath.Join(scanDir, "dns_reverse_lookup.txt"),
					filepath.Join(scanDir, "windows_os_discovery.txt"),
					filepath.Join(scanDir, "nmap_discovery_scan.xml"),
				)
			},
			critical: false, // Set to true if categorization is critical
		},
	}

	// Perform all scans
	for _, step := range scanSteps {
		if step.critical {
			if err := step.execute(); err != nil {

				(*outputModal).AppendText(fmt.Sprintf("[red]%s failed: %v[-]\n", step.name, err))
				logger.Error("%s failed: %v", step.name, err)
				(*outputModal).SetScanning(false)
				(*outputModal).CloseOutputModal(fmt.Sprintf("%s Failed", step.name))

				return
			}
		} else {
			if err := step.execute(); err != nil {
				(*outputModal).AppendText(fmt.Sprintf("[yellow]%s failed: %v[-]\n", step.name, err))
				logger.Error("%s failed: %v", step.name, err)
			}
		}
	}

	// All scans completed successfully
	close(cleanupDone) // This will trigger the cleanup goroutine to close the modal
}
