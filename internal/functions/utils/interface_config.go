package utils

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// GetAllConfiguredInterfaces returns a list of all interfaces including VLANs
func GetAllConfiguredInterfaces(mainInterface string) ([]string, error) {
	interfaces := []string{mainInterface}

	// Get all system interfaces
	allInterfaces, err := GetSubinterfaces(mainInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	// Find subinterfaces of the main interface
	for _, iface := range allInterfaces {
		// Check if this is a subinterface of our main interface
		if strings.HasPrefix(iface, mainInterface+".") {
			interfaces = append(interfaces, iface)
		}
	}

	return interfaces, nil
}

// ConfigureVLANs creates and configures VLAN interfaces
func ConfigureVLANs(mainInterface string, vlanIDs []string) error {
	for _, vlanID := range vlanIDs {
		vlanName := fmt.Sprintf("%s.%s", mainInterface, vlanID)

		// Create VLAN interface
		cmd := exec.Command("ip", "link", "add", "link", mainInterface,
			"name", vlanName, "type", "vlan", "id", vlanID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create VLAN interface %s: %v", vlanName, err)
		}

		// Bring up the interface
		cmd = exec.Command("ip", "link", "set", vlanName, "up")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to bring up VLAN interface %s: %v", vlanName, err)
		}
	}
	return nil
}

// ConfigureIPAddresses configures IP addresses for the selected interfaces
func ConfigureIPAddresses(interfaces []string, app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	cfg, err := configuration.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if cfg.NetworkInterfaces == nil {
		cfg.NetworkInterfaces = make(map[string]configuration.InterfaceState)
	}

	for _, iface := range interfaces {
		// Prompt for IP address configuration
		uiutil.PromptInput(app, pages, fmt.Sprintf("ipConfigModal-%s", iface),
			fmt.Sprintf("Configure IP for %s", iface),
			"Enter IP address with subnet mask (e.g., 192.168.1.1/24):",
			mainView,
			func(input string, err error) {
				if err != nil {
					return
				}

				// Validate IP address format
				ip, ipNet, err := net.ParseCIDR(input)
				if err != nil {
					logger.Error("Invalid IP address format for %s: %v", iface, err)
					uiutil.ShowError(app, pages, "invalidIPFormatModal",
						"Invalid IP format. Please use CIDR notation (e.g., 192.168.1.1/24)",
						mainView, nil)
					return
				}

				// Apply IP configuration
				err = applyIPConfig(iface, input)
				if err != nil {
					logger.Error("Failed to apply IP configuration for %s: %v", iface, err)
					uiutil.ShowError(app, pages, "applyIPConfigErrorModal",
						fmt.Sprintf("Failed to configure IP for %s: %v", iface, err),
						mainView, nil)
					return
				}

				// Update configuration
				updateInterfaceConfig(cfg, iface, ip.String(), ipNet.String())
				err = configuration.SaveConfig(cfg)
				if err != nil {
					logger.Error("Failed to save configuration: %v", err)
					uiutil.ShowError(app, pages, "saveConfigErrorModal",
						fmt.Sprintf("Failed to save configuration: %v", err),
						mainView, nil)
					return
				}

				logger.Info("Successfully configured IP %s for interface %s", input, iface)
				uiutil.ShowMessage(app, pages, "ipConfigSuccessModal",
					fmt.Sprintf("Successfully configured IP for %s", iface),
					mainView)
			},
			"")
	}

	return nil
}

// applyIPConfig applies the IP configuration to the interface
func applyIPConfig(iface, ip string) error {
	// First, flush existing IP addresses
	flushCmd := exec.Command("ip", "addr", "flush", "dev", iface)
	if err := flushCmd.Run(); err != nil {
		logger.Error("failed to flush IP addresses: %v", err)
		return fmt.Errorf("failed to flush IP addresses: %v", err)
	}

	// Apply new IP address
	addCmd := exec.Command("ip", "addr", "add", ip, "dev", iface)
	if err := addCmd.Run(); err != nil {
		logger.Error("failed to add IP address: %v", err)
		return fmt.Errorf("failed to add IP address: %v", err)
	}

	return nil
}

// updateInterfaceConfig updates the configuration with new IP settings
func updateInterfaceConfig(cfg *configuration.Config, iface, ip, subnet string) {
	parts := strings.Split(iface, ".")
	mainIface := parts[0]

	if len(parts) == 1 {
		// Main interface
		cfg.NetworkInterfaces[mainIface] = configuration.InterfaceState{
			Status:     "up",
			IPAddress:  ip,
			SubnetMask: subnet,
			LinkState:  "up",
		}
	} else {
		// VLAN interface
		mainIfaceState := cfg.NetworkInterfaces[mainIface]
		if mainIfaceState.Subinterfaces == nil {
			mainIfaceState.Subinterfaces = []configuration.SubinterfaceState{}
		}

		// Update or add subinterface
		found := false
		for i, sub := range mainIfaceState.Subinterfaces {
			if sub.Name == iface {
				mainIfaceState.Subinterfaces[i].IPAddress = ip
				mainIfaceState.Subinterfaces[i].SubnetMask = subnet
				found = true
				break
			}
		}

		if !found {
			mainIfaceState.Subinterfaces = append(mainIfaceState.Subinterfaces,
				configuration.SubinterfaceState{
					Name:       iface,
					IPAddress:  ip,
					SubnetMask: subnet,
				})
		}

		cfg.NetworkInterfaces[mainIface] = mainIfaceState
	}
}
