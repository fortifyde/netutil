package functions

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/functions/utils"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// start wireshark and tshark captures, perform analysis, and update the UI accordingly
func StartWiresharkListening(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Starting Wireshark Listening")

	// check for root access
	if os.Geteuid() != 0 {
		logger.Error("Root access required for Wireshark listening")
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "rootAccessRequiredModal", "Root access required for Wireshark listening. Please run the program with sudo.", mainView, nil)
		})
		return fmt.Errorf("root access required")
	}

	requiredBinaries := []string{"wireshark", "tshark"}
	if !utils.CheckDependencies(requiredBinaries) {
		msg := "Required binaries not found: " + strings.Join(requiredBinaries, ", ")
		logger.Error("Required binaries not found: %s", strings.Join(requiredBinaries, ", "))
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "requiredBinariesNotFoundModal", msg, mainView, nil)
		})
		return errors.New(msg)
	}
	cfg, err := configuration.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "loadConfigErrorModal", fmt.Sprintf("Failed to load config: %v", err), mainView, nil)
		})
		return fmt.Errorf("failed to load config: %v", err)
	}

	captureDir := filepath.Join(cfg.WorkingDirectory, "Wireshark")
	analysisDir := filepath.Join(captureDir, "capture_analysis")

	// ensure capture and analysisdirectories exist
	if err := utils.EnsureDir(captureDir); err != nil {
		logger.Error("Failed to create capture directory: %v", err)
		return err
	}
	if err := utils.EnsureDir(analysisDir); err != nil {
		logger.Error("Failed to create analysis directory: %v", err)
		return err
	}
	// get available interfaces
	interfaces, err := utils.GetEthernetInterfaces()
	if err != nil {
		logger.Error("Failed to get Ethernet interfaces: %v", err)
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "getEthernetInterfacesErrorModal", fmt.Sprintf("Failed to get Ethernet interfaces: %v", err), mainView, nil)
		})
		return fmt.Errorf("failed to get Ethernet interfaces: %v", err)
	}

	// filter active interfaces
	var activeInterfaces []net.Interface
	for _, iface := range interfaces {
		status, _ := utils.GetInterfaceStatus(iface.Name)
		if status == "up" {
			activeInterfaces = append(activeInterfaces, iface)
		}
	}

	if len(activeInterfaces) == 0 {
		logger.Error("No active interfaces found")
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "noActiveInterfacesFoundModal", "No active interfaces found", mainView, nil)
		})
		return fmt.Errorf("no active interfaces found")
	}

	var selectedInterface string
	if len(activeInterfaces) == 1 {
		selectedInterface = activeInterfaces[0].Name
	} else {
		// show interface selection dialog if more than one interface is active
		interfaceNames := make([]string, len(activeInterfaces))
		for i, iface := range activeInterfaces {
			interfaceNames[i] = iface.Name
		}

		uiutil.ShowList(app, pages, "selectInterfaceForWiresharkModal", "Select Interface for Wireshark", interfaceNames, func(index int) {
			selectedInterface = activeInterfaces[index].Name
		}, mainView)

		// wait for the user to select an interface
		if selectedInterface == "" {
			logger.Error("No interface selected")
			app.QueueUpdateDraw(func() {
				uiutil.ShowError(app, pages, "noInterfaceSelectedModal", "No interface selected", mainView, nil)
			})
			return fmt.Errorf("no interface selected")
		}
	}

	// create output files
	timestamp := time.Now().Format("20060102_150405")
	wiresharkOutputFile := filepath.Join(captureDir, fmt.Sprintf("wireshark_capture_%s.pcapng", timestamp))
	tsharkOutputFile := filepath.Join(captureDir, fmt.Sprintf("tshark_capture_%s.pcap", timestamp))

	// start wireshark for live viewing
	wiresharkCmd := exec.Command("wireshark",
		"-i", selectedInterface,
		"-k",
		"-w", wiresharkOutputFile,
		"-n",
		"-l",
		"-S",
		"-a", "duration:600",
		"-a", "filesize:5000000",
	)
	wiresharkCmd.Stderr = io.Discard
	err = wiresharkCmd.Start()
	if err != nil {
		logger.Error("Failed to start Wireshark: %v", err)
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "startWiresharkErrorModal", "Failed to start Wireshark.", mainView, nil)
		})
		return err
	}

	// start tshark for capturing without GUI
	tsharkCmd := exec.Command("tshark",
		"-i", selectedInterface,
		"-w", tsharkOutputFile,
		"-n",
		"-a", "duration:10",
		"-a", "filesize:5000000",
	)
	tsharkCmd.Stderr = io.Discard
	err = tsharkCmd.Start()
	if err != nil {
		logger.Error("Failed to start tshark: %v", err)
		if err := wiresharkCmd.Process.Kill(); err != nil {
			logger.Error("Failed to kill wireshark process: %v", err)
		}
		app.QueueUpdateDraw(func() {
			uiutil.ShowError(app, pages, "startTsharkErrorModal", "Failed to start tshark.", mainView, nil)
		})
		return err
	}

	logger.Debug("Starting tshark with command: %v", tsharkCmd.Args)

	go func() {
		// wait for tshark to finish
		err := tsharkCmd.Wait()
		if err != nil {
			logger.Error("tshark capture ended with error: %v", err)
			app.QueueUpdateDraw(func() {
				uiutil.ShowError(app, pages, "tsharkCaptureErrorModal", fmt.Sprintf("tshark capture ended with error: %v", err), mainView, nil)
			})
			return
		}

		// perform analysis
		results, err := AnalyzeTsharkCapture(tsharkOutputFile, analysisDir) // analysisDir is now inside captureDir
		if err != nil {
			logger.Error("Analysis failed: %v", err)
			app.QueueUpdateDraw(func() {
				uiutil.ShowError(app, pages, "analysisFailedModal", fmt.Sprintf("Analysis failed: %v", err), mainView, nil)
			})
			return
		}

		// ask user to view results
		app.QueueUpdateDraw(func() {
			uiutil.ShowConfirm(app, pages, "captureAndAnalysisCompletedModal", "Capture and analysis completed and saved in Wireshark folder. Would you like to see the most relevant data?", func(yes bool) {
				if yes {
					uiutil.ShowAnalysisResults(app, pages, "analysisResultsModal", results, mainView)
				} else {
					uiutil.ShowMessage(app, pages, "analysisCompletedModal", "Analysis completed. Results saved in the 'Wireshark/capture_analysis' folder.", mainView)
				}
			}, mainView)
		})

		// attemt to gracefully close wireshark. after 2 seconds, kill process
		if err := wiresharkCmd.Process.Signal(os.Interrupt); err != nil {
			logger.Warning("Failed to send interrupt to Wireshark: %v", err)
		}
		time.Sleep(2 * time.Second)
		if err := wiresharkCmd.Process.Kill(); err != nil {
			logger.Critical("Failed to kill Wireshark process: %v", err)
		}

	}()

	return nil
}
