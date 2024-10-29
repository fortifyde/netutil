package utils

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

var (
	mainInterface    string
	mainInterfaceMux sync.RWMutex
)

// GetMainInterface returns the currently set main interface
func GetMainInterface() string {
	mainInterfaceMux.RLock()
	defer mainInterfaceMux.RUnlock()

	if mainInterface == "" {
		// Try to load from config
		if cfg, err := configuration.LoadConfig(); err == nil {
			mainInterface = cfg.MainInterface
		}
	}
	return mainInterface
}

// SetMainInterface sets the main interface
func SetMainInterface(iface string) error {
	mainInterfaceMux.Lock()
	defer mainInterfaceMux.Unlock()
	mainInterface = iface

	// Save to configuration
	cfg, err := configuration.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	cfg.MainInterface = iface
	if err := configuration.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

// ConfigureNetworkInterfaces handles the configuration of the main network interface
func ConfigureNetworkInterfaces(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Configuring network interfaces")

	// Get active ethernet interfaces
	activeInterfaces, err := GetEthernetInterfaces()
	if err != nil {
		logger.Error("Failed to get ethernet interfaces: %v", err)
		return fmt.Errorf("failed to get ethernet interfaces: %v", err)
	}

	// Filter for active interfaces
	var availableInterfaces []net.Interface
	for _, iface := range activeInterfaces {
		if status, err := GetInterfaceStatus(iface.Name); err == nil && status == "up" {
			availableInterfaces = append(availableInterfaces, iface)
		}
	}

	if len(availableInterfaces) == 0 {
		const errMsg = "No active ethernet interfaces found!"
		logger.Error(errMsg)
		uiutil.ShowError(app, pages, "noActiveInterfacesModal", errMsg, mainView, nil)
		return errors.New("no active interfaces found")
	}

	// If main interface is already set, show current configuration
	if GetMainInterface() != "" {
		showCurrentConfig(app, pages, mainView)
		return nil
	}

	// If only one active interface, set it automatically
	if len(availableInterfaces) == 1 {
		if err := SetMainInterface(availableInterfaces[0].Name); err != nil {
			logger.Error("Failed to set main interface: %v", err)
			return err
		}
		msg := fmt.Sprintf("Main interface automatically set to %s", availableInterfaces[0].Name)
		logger.Info(msg)
		uiutil.ShowMessage(app, pages, "mainInterfaceSetModal", msg, mainView)
		return nil
	}

	// Show interface selection dialog
	interfaceNames := make([]string, len(availableInterfaces))
	for i, iface := range availableInterfaces {
		interfaceNames[i] = iface.Name
	}

	uiutil.ShowList(app, pages, "selectMainInterfaceModal",
		"Select Main Network Interface",
		interfaceNames,
		func(index int) {
			if index >= 0 && index < len(availableInterfaces) {
				if err := SetMainInterface(availableInterfaces[index].Name); err != nil {
					logger.Error("Failed to set main interface: %v", err)
					uiutil.ShowError(app, pages, "setMainInterfaceErrorModal",
						fmt.Sprintf("Failed to set main interface: %v", err), mainView, nil)
					return
				}
				msg := fmt.Sprintf("Main interface set to %s", availableInterfaces[index].Name)
				logger.Info(msg)
			}
		},
		mainView)

	return nil
}

func showCurrentConfig(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) {
	currentInterface := GetMainInterface()
	msg := fmt.Sprintf("Current main interface: %s\nWould you like to change it?", currentInterface)

	uiutil.ShowConfirm(app, pages, "changeMainInterfaceModal", msg,
		func(confirmed bool) {
			if confirmed {
				// Reset the main interface and recall the configuration function
				if err := SetMainInterface(""); err != nil {
					logger.Error("Failed to reset main interface: %v", err)
					return
				}
				if err := ConfigureNetworkInterfaces(app, pages, mainView); err != nil {
					logger.Error("Failed to reconfigure network interfaces: %v", err)
					return
				}
			}
		},
		mainView)
}
