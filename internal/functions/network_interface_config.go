package functions

import (
	"fmt"
	"strings"

	"github.com/fortifyde/netutil/internal/functions/utils"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

func ConfigureNetworkInterfaces(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	// First, ensure we have a main interface set
	mainInterface := utils.GetMainInterface()
	if mainInterface == "" {
		err := utils.ConfigureNetworkInterfaces(app, pages, mainView)
		if err != nil {
			return fmt.Errorf("failed to configure main interface: %v", err)
		}
		mainInterface = utils.GetMainInterface()
	}

	// Initial choice between VLAN or IP configuration
	options := []string{
		"Configure VLANs",
		"Configure IP Addresses",
		"Cancel",
	}

	uiutil.ShowList(app, pages, "networkConfigTypeModal",
		fmt.Sprintf("Network Configuration for %s", mainInterface),
		options,
		func(index int) {
			switch index {
			case 0:
				showVLANConfigOptions(app, pages, mainInterface, mainView)
			case 1:
				showIPConfigOptions(app, pages, mainInterface, mainView)
			case 2:
				return
			}
		},
		mainView)

	return nil
}

func showVLANConfigOptions(app *tview.Application, pages *tview.Pages, mainInterface string, mainView tview.Primitive) {
	options := []string{
		"Automatic (from Wireshark results)",
		"Manual configuration",
		"Cancel",
	}

	uiutil.ShowList(app, pages, "vlanConfigTypeModal",
		"VLAN Configuration Method",
		options,
		func(index int) {
			switch index {
			case 0:
				configureVLANsFromWireshark(app, pages, mainInterface, mainView)
			case 1:
				configureVLANsManually(app, pages, mainInterface, mainView)
			case 2:
				return
			}
		},
		mainView)
}

func configureVLANsFromWireshark(app *tview.Application, pages *tview.Pages, mainInterface string, mainView tview.Primitive) {
	// Get VLANs from Wireshark results
	vlanIDs, err := utils.GetWiresharkVLANs()
	if err != nil {
		logger.Error("Failed to get VLANs from Wireshark: %v", err)
		uiutil.ShowError(app, pages, "getWiresharkVLANsErrorModal",
			fmt.Sprintf("Failed to get VLANs from Wireshark: %v", err),
			mainView, nil)
		return
	}

	if len(vlanIDs) == 0 {
		uiutil.ShowMessage(app, pages, "noVLANsFoundModal",
			"No VLANs found in Wireshark results",
			mainView)
		return
	}

	showVLANSelectionList(app, pages, mainInterface, vlanIDs, mainView)
}

func configureVLANsManually(app *tview.Application, pages *tview.Pages, mainInterface string, mainView tview.Primitive) {
	uiutil.PromptInput(app, pages, "manualVLANInputModal",
		"Manual VLAN Configuration",
		"Enter VLAN IDs (separated by space):",
		mainView,
		func(input string, err error) {
			if err != nil {
				return
			}
			vlanIDs := strings.Fields(input)
			if len(vlanIDs) == 0 {
				uiutil.ShowError(app, pages, "noVLANsEnteredModal",
					"No VLAN IDs entered",
					mainView, nil)
				return
			}
			utils.ConfigureVLANs(mainInterface, vlanIDs)
		},
		"")
}

func showIPConfigOptions(app *tview.Application, pages *tview.Pages, mainInterface string, mainView tview.Primitive) {
	// Get all configured interfaces including VLANs
	interfaces, err := utils.GetAllConfiguredInterfaces(mainInterface)
	if err != nil {
		logger.Error("Failed to get configured interfaces: %v", err)
		uiutil.ShowError(app, pages, "getInterfacesErrorModal",
			fmt.Sprintf("Failed to get configured interfaces: %v", err),
			mainView, nil)
		return
	}

	showInterfaceSelectionList(app, pages, interfaces, mainView)
}

func showVLANSelectionList(app *tview.Application, pages *tview.Pages, mainInterface string, vlanIDs []string, mainView tview.Primitive) {
	uiutil.ShowMultiSelect(app, pages, "vlanSelectionModal",
		fmt.Sprintf("Select VLANs for %s", mainInterface),
		vlanIDs,
		func(selected []string) {
			if len(selected) > 0 {
				utils.ConfigureVLANs(mainInterface, selected)
			}
		},
		mainView)
}

func showInterfaceSelectionList(app *tview.Application, pages *tview.Pages, interfaces []string, mainView tview.Primitive) {
	uiutil.ShowMultiSelect(app, pages, "interfaceSelectionModal",
		"Select Interfaces to Configure",
		interfaces,
		func(selected []string) {
			if len(selected) > 0 {
				err := utils.ConfigureIPAddresses(selected, app, pages, mainView)
				if err != nil {
					logger.Error("Failed to configure IP addresses: %v", err)
					uiutil.ShowError(app, pages, "configureIPErrorModal",
						fmt.Sprintf("Failed to configure IP addresses: %v", err),
						mainView, nil)
				}
			}
		},
		mainView)
}
