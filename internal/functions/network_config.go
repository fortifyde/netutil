package functions

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

// Function for saving and loading network configurations.
// It retrieves the list of Ethernet interfaces (including subinterfaces),
// their IP configurations, and the default route.
//
// Requires root access to modify network configurations.

func checkElevatedAccess() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("root access required")
	}
	return nil
}

func getDefaultRoute() (string, error) {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default route: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "default" {
			return fields[2], nil // Return the gateway IP address
		}
	}

	return "", fmt.Errorf("no default route found")
}

func getInterfaceIPConfig(ifaceName string) (string, string, error) {
	cmd := exec.Command("ip", "-o", "-4", "addr", "show", "dev", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get IP config: %v", err)
	}

	re := regexp.MustCompile(`inet\s+(\d+\.\d+\.\d+\.\d+)/(\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 3 {
		return "", "", fmt.Errorf("IP address not found for interface %s", ifaceName)
	}

	ipAddress := matches[1]
	cidr := matches[2]
	subnetMask := cidrToSubnetMask(cidr)

	return ipAddress, subnetMask, nil
}

func getInterfaceLinkState(ifaceName string) (string, error) {
	cmd := exec.Command("ip", "link", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get link state: %v", err)
	}

	if strings.Contains(string(output), "state UP") {
		return "up", nil
	} else if strings.Contains(string(output), "state DOWN") {
		return "down", nil
	}

	return "unknown", nil
}

func applyInterfaceConfig(ifaceName string, state InterfaceState) error {
	logger.Info("Applying configuration for interface: %s", ifaceName)
	if err := checkElevatedAccess(); err != nil {
		return err
	}

	// Set IP address and subnet mask
	cmd := exec.Command("ip", "addr", "flush", "dev", ifaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush IP address: %v", err)
	}

	cmd = exec.Command("ip", "addr", "add", state.IPAddress+"/"+subnetMaskToCIDR(state.SubnetMask), "dev", ifaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set IP address: %v", err)
	}

	// Set link state
	linkCmd := "up"
	if state.LinkState == "down" {
		linkCmd = "down"
	}
	cmd = exec.Command("ip", "link", "set", ifaceName, linkCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set link state: %v", err)
	}

	// Configure subinterfaces
	for _, sub := range state.Subinterfaces {
		vlanID := strings.TrimPrefix(sub.Name, ifaceName+".")
		cmd = exec.Command("ip", "link", "add", "link", ifaceName, "name", sub.Name, "type", "vlan", "id", vlanID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create subinterface %s: %v", sub.Name, err)
		}

		cmd = exec.Command("ip", "addr", "add", sub.IPAddress+"/"+subnetMaskToCIDR(sub.SubnetMask), "dev", sub.Name)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set IP for subinterface %s: %v", sub.Name, err)
		}

		cmd = exec.Command("ip", "link", "set", sub.Name, "up")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to bring up subinterface %s: %v", sub.Name, err)
		}
	}

	return nil
}

func cidrToSubnetMask(cidr string) string {
	bits, _ := strconv.Atoi(cidr)
	mask := net.CIDRMask(bits, 32)
	return net.IP(mask).String()
}

func subnetMaskToCIDR(subnetMask string) string {
	ipv4Mask := net.IPMask(net.ParseIP(subnetMask).To4())
	ones, _ := ipv4Mask.Size()
	return strconv.Itoa(ones)
}

func SaveNetworkConfig(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Saving network configuration")
	interfaces, err := GetEthernetInterfaces()
	if err != nil {
		logger.Error("Failed to get Ethernet interfaces: %v", err)
		return fmt.Errorf("failed to get Ethernet interfaces: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return fmt.Errorf("failed to load config: %v", err)
	}

	cfg.NetworkInterfaces = make(map[string]InterfaceState)
	// get default route
	defaultRoute, err := getDefaultRoute()
	if err != nil {
		uiutil.ShowError(app, pages, fmt.Sprintf("Failed to get default route: %v", err), mainView)
	} else {
		cfg.DefaultRoute = defaultRoute
	}
	// get interface states and save IPAddress and SubnetMask if up.
	// also save subinterface states if any.
	for _, iface := range interfaces {
		linkState, err := getInterfaceLinkState(iface.Name)
		if err != nil {
			return fmt.Errorf("failed to get link state for interface %s: %v", iface.Name, err)
		}

		// only save interfaces that are up
		if linkState == "up" {
			status, err := getInterfaceStatus(iface.Name)
			if err != nil {
				return fmt.Errorf("failed to get status for interface %s: %v", iface.Name, err)
			}

			ipAddress, subnetMask, err := getInterfaceIPConfig(iface.Name)
			if err != nil {
				return fmt.Errorf("failed to get IP config for interface %s: %v", iface.Name, err)
			}

			subinterfaceNames, err := GetSubinterfaces(iface.Name)
			if err != nil {
				return fmt.Errorf("failed to get subinterfaces for interface %s: %v", iface.Name, err)
			}

			var subinterfaces []SubinterfaceState
			for _, subName := range subinterfaceNames {
				subIP, subMask, err := getInterfaceIPConfig(subName)
				if err != nil {
					return fmt.Errorf("failed to get IP config for subinterface %s: %v", subName, err)
				}
				subinterfaces = append(subinterfaces, SubinterfaceState{
					Name:       subName,
					IPAddress:  subIP,
					SubnetMask: subMask,
				})
			}

			cfg.NetworkInterfaces[iface.Name] = InterfaceState{
				Status:        status,
				IPAddress:     ipAddress,
				SubnetMask:    subnetMask,
				LinkState:     linkState,
				Subinterfaces: subinterfaces,
			}
		}
	}

	err = SaveConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	logger.Info("Network configuration saved successfully")
	uiutil.ShowMessage(app, pages, "Network configuration saved successfully", mainView)
	return nil
}

// main function to apply a saved config to the system
func LoadAndApplyNetworkConfig(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Loading and applying network configuration")
	if err := checkElevatedAccess(); err != nil {
		logger.Error("Elevated access required to apply network config: %v", err)
		return fmt.Errorf("elevated access required to apply network config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if len(cfg.NetworkInterfaces) == 0 {
		uiutil.ShowMessage(app, pages, "No saved network configuration found", mainView)
		return nil
	}

	for ifaceName, state := range cfg.NetworkInterfaces {
		err := applyInterfaceConfig(ifaceName, state)
		if err != nil {
			uiutil.ShowError(app, pages, fmt.Sprintf("Failed to apply config for interface %s: %v", ifaceName, err), mainView)
		} else {
			uiutil.ShowMessage(app, pages, fmt.Sprintf("Configuration applied for interface %s", ifaceName), mainView)
		}
	}

	if cfg.DefaultRoute != "" {
		err := applyDefaultRoute(cfg.DefaultRoute)
		if err != nil {
			uiutil.ShowError(app, pages, fmt.Sprintf("Failed to apply default route: %v", err), mainView)
		} else {
			uiutil.ShowMessage(app, pages, "Default route applied successfully", mainView)
		}
	}

	logger.Info("Network configuration applied successfully")
	uiutil.ShowMessage(app, pages, "Network configuration applied successfully", mainView)
	return nil
}

func applyDefaultRoute(gateway string) error {
	logger.Info("Applying default route: %s", gateway)
	cmd := exec.Command("ip", "route", "replace", "default", "via", gateway)
	return cmd.Run()
}
