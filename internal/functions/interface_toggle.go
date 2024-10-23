package functions

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// functions to handle toggling of ethernet interfaces
// retrieve all ethernet interfaces and either toggle a single interface or
// show a list of interfaces to choose from
// toggling a single enabled interface will prompt for confirmation to disable it

func ToggleEthernetInterfaces(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) {
	logger.Info("Toggling Ethernet interfaces")
	interfaces, err := GetEthernetInterfaces()
	if err != nil {
		logger.Error("Failed to get Ethernet interfaces: %v", err)
		uiutil.ShowError(app, pages, "Error getting interfaces: "+err.Error(), mainView, nil)
		return
	}

	if len(interfaces) == 0 {
		logger.Info("No Ethernet interfaces found")
		uiutil.ShowMessage(app, pages, "No Ethernet interfaces found.", mainView)
		return
	}

	if len(interfaces) == 1 {
		toggleSingleInterface(app, pages, interfaces[0], mainView)
	} else {
		showInterfaceList(app, pages, interfaces, mainView)
	}
}

func toggleSingleInterface(app *tview.Application, pages *tview.Pages, iface net.Interface, mainView tview.Primitive) {
	logger.Info("Attempting to toggle single interface: %s", iface.Name)
	status, err := getInterfaceStatus(iface.Name)
	if err != nil {
		logger.Error("Failed to get status for interface %s: %v", iface.Name, err)
		uiutil.ShowError(app, pages, "Error getting interface status: "+err.Error(), mainView, nil)
		return
	}

	if status == "down" {
		logger.Info("Attempting to enable interface %s", iface.Name)
		err = setInterfaceStatus(iface.Name, "up")
		if err != nil {
			logger.Error("Failed to enable interface %s: %v", iface.Name, err)
			uiutil.ShowError(app, pages, "Error enabling interface: "+err.Error(), mainView, nil)
		} else {
			logger.Info("Successfully enabled interface %s", iface.Name)
			uiutil.ShowMessage(app, pages, fmt.Sprintf("Interface %s has been enabled.", iface.Name), mainView)
		}
	} else {
		confirmDisable(app, pages, iface.Name, mainView)
	}
}

func showInterfaceList(app *tview.Application, pages *tview.Pages, interfaces []net.Interface, mainView tview.Primitive) {
	items := make([]string, len(interfaces)+1)
	for i, iface := range interfaces {
		status, _ := getInterfaceStatus(iface.Name)
		items[i] = fmt.Sprintf("%s (%s)", iface.Name, status)
	}
	items[len(interfaces)] = "Cancel"

	uiutil.ShowList(app, pages, "Select Ethernet Interface to Toggle", items, func(index int) {
		if index < len(interfaces) {
			toggleInterface(app, pages, interfaces[index].Name, mainView)
		}
	}, mainView)
}

func toggleInterface(app *tview.Application, pages *tview.Pages, name string, mainView tview.Primitive) {
	logger.Info("Attempting to toggle interface: %s", name)
	status, err := getInterfaceStatus(name)
	if err != nil {
		logger.Error("Failed to get status for interface %s: %v", name, err)
		uiutil.ShowError(app, pages, "Error getting interface status: "+err.Error(), mainView, nil)
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
		uiutil.ShowError(app, pages, fmt.Sprintf("Error setting interface %s to %s: %s\n\n[red]root access needed!\n", name, newStatus, err.Error()), mainView, nil)
	} else {
		logger.Info("Successfully set interface %s to %s", name, newStatus)
		uiutil.ShowMessage(app, pages, fmt.Sprintf("Interface %s has been set to %s.", name, newStatus), mainView)
	}
}

func confirmDisable(app *tview.Application, pages *tview.Pages, name string, mainView tview.Primitive) {
	logger.Info("Confirming disable for interface: %s", name)
	uiutil.ShowConfirm(app, pages, fmt.Sprintf("Do you want to disable interface %s?", name), func(confirmed bool) {
		if confirmed {
			logger.Info("User confirmed disabling interface %s", name)
			err := setInterfaceStatus(name, "down")
			if err != nil {
				logger.Error("Failed to disable interface %s: %v", name, err)
				uiutil.ShowError(app, pages, "Error disabling interface: "+err.Error(), mainView, nil)
			} else {
				logger.Info("Successfully disabled interface %s", name)
				uiutil.ShowMessage(app, pages, fmt.Sprintf("Interface %s has been disabled.", name), mainView)
			}
		} else {
			logger.Info("User cancelled disabling interface %s", name)
		}
	}, mainView)
}

func getInterfaceStatus(name string) (string, error) {
	logger.Info("Getting status for interface: %s", name)
	cmd := exec.Command("ip", "link", "show", name)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("Failed to get status for interface %s: %v", name, err)
		return "", err
	}

	if strings.Contains(string(output), "state UP") {
		return "up", nil
	}
	return "down", nil
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
