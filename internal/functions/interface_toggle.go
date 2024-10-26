package functions

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/fortifyde/netutil/internal/functions/utils"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// functions to handle toggling of ethernet interfaces
// retrieve all ethernet interfaces and either toggle a single interface or
// show a list of interfaces to choose from
// toggling a single enabled interface will prompt for confirmation to disable it

func ToggleEthernetInterfaces(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Toggling Ethernet interfaces")
	interfaces, err := utils.GetEthernetInterfaces()
	if err != nil {
		logger.Error("Failed to get Ethernet interfaces: %v", err)
		uiutil.ShowError(app, pages, "getEthernetInterfacesErrorModal", "Error getting interfaces: "+err.Error(), mainView, nil)
		return err
	}

	if len(interfaces) == 0 {
		logger.Info("No Ethernet interfaces found")
		uiutil.ShowMessage(app, pages, "noEthernetInterfacesFoundModal", "No Ethernet interfaces found.", mainView)
		return err
	}

	if len(interfaces) == 1 {
		toggleSingleInterface(app, pages, interfaces[0], mainView)
	} else {
		showInterfaceList(app, pages, interfaces, mainView)
	}
	return nil
}

func toggleSingleInterface(app *tview.Application, pages *tview.Pages, iface net.Interface, mainView tview.Primitive) {
	currentStatus, err := utils.GetInterfaceStatus(iface.Name)
	if err != nil {
		logger.Error("Failed to get status for interface %s: %v", iface.Name, err)
		uiutil.ShowError(app, pages, "getInterfaceStatusErrorModal", fmt.Sprintf("Error: %v", err), mainView, nil)
		return
	}

	newStatus := "down"
	if currentStatus == "down" {
		newStatus = "up"
	}

	uiutil.ShowConfirm(app, pages, "toggleInterfaceConfirmModal", fmt.Sprintf("Are you sure you want to set interface %s to %s?", iface.Name, newStatus), func(yes bool) {
		if yes {
			err := setInterfaceStatus(iface.Name, newStatus)
			if err != nil {
				logger.Error("Failed to set interface %s to %s: %v", iface.Name, newStatus, err)
				uiutil.ShowError(app, pages, "setInterfaceStatusErrorModal", fmt.Sprintf("Error: %v", err), mainView, nil)
				return
			}
			logger.Info("Interface %s set to %s", iface.Name, newStatus)
			uiutil.ShowMessage(app, pages, "setInterfaceStatusSuccessModal", fmt.Sprintf("Interface %s set to %s.", iface.Name, newStatus), mainView)
		} else {
			logger.Info("User canceled toggling interface %s", iface.Name)
		}
	}, mainView)
}

func showInterfaceList(app *tview.Application, pages *tview.Pages, interfaces []net.Interface, mainView tview.Primitive) {
	logger.Info("Starting showInterfaceList")
	items := make([]string, len(interfaces)+1)
	for i, iface := range interfaces {
		status, _ := utils.GetInterfaceStatus(iface.Name)
		items[i] = fmt.Sprintf("%s (currently %s)", iface.Name, status)
		logger.Info("Added interface to list: %s", items[i])
	}
	items[len(interfaces)] = "Cancel"

	logger.Info("About to call ShowList, modal active: %v", uiutil.GetFloatingBoxActive())
	uiutil.ShowList(app, pages, "selectEthernetInterfaceToToggleModal", "Select Ethernet Interface to Toggle", items, func(index int) {
		logger.Info("List item selected: %d", index)
		if index < len(interfaces) {
			toggleInterface(app, pages, interfaces[index].Name, mainView)
		}
	}, mainView)
	logger.Info("ShowList called")
}

func toggleInterface(app *tview.Application, pages *tview.Pages, name string, mainView tview.Primitive) {
	logger.Info("Attempting to toggle interface: %s", name)
	status, err := utils.GetInterfaceStatus(name)
	if err != nil {
		logger.Error("Failed to get status for interface %s: %v", name, err)
		uiutil.ShowError(app, pages, "getInterfaceStatusErrorModal", "Error getting interface status: "+err.Error(), mainView, nil)
		return
	}

	newStatus := "up"
	if status == "up" {
		newStatus = "down"
	}

	logger.Info("Attempting to set interface %s to %s", name, newStatus)
	err = setInterfaceStatus(name, newStatus)
	if err != nil {
		logger.Error("Failed to set interface %s to %s: %v", name, newStatus, err)
		uiutil.ShowError(app, pages, "setInterfaceStatusErrorModal", fmt.Sprintf("Error setting interface %s to %s: %s\n\n[red]root access needed!\n", name, newStatus, err.Error()), mainView, nil)
	} else {
		logger.Info("Successfully set interface %s to %s", name, newStatus)
		uiutil.ShowMessage(app, pages, "setInterfaceStatusSuccessModal", fmt.Sprintf("Interface %s has been set to %s.", name, newStatus), mainView)
	}
}

func setInterfaceStatus(name, status string) error {
	logger.Info("Setting interface %s status to %s", name, status)
	cmd := exec.Command("ip", "link", "set", name, status)
	err := cmd.Run()
	if err != nil {
		logger.Error("Failed to set interface %s status to %s: %v", name, status, err)
	}
	return err
}
