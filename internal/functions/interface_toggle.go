package functions

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func ToggleEthernetInterfaces(app *tview.Application, mainView tview.Primitive) {
	interfaces, err := GetEthernetInterfaces()
	if err != nil {
		uiutil.ShowError(app, "Error getting interfaces: "+err.Error(), mainView)
		return
	}

	if len(interfaces) == 0 {
		uiutil.ShowMessage(app, "No Ethernet interfaces found.", mainView)
		return
	}

	if len(interfaces) == 1 {
		toggleSingleInterface(app, interfaces[0], mainView)
	} else {
		showInterfaceList(app, interfaces, mainView)
	}
}

func toggleSingleInterface(app *tview.Application, iface net.Interface, mainView tview.Primitive) {
	status, err := getInterfaceStatus(iface.Name)
	if err != nil {
		uiutil.ShowError(app, "Error getting interface status: "+err.Error(), mainView)
		return
	}

	if status == "down" {
		err = setInterfaceStatus(iface.Name, "up")
		if err != nil {
			uiutil.ShowError(app, "Error enabling interface: "+err.Error(), mainView)
		} else {
			uiutil.ShowMessage(app, fmt.Sprintf("Interface %s has been enabled.", iface.Name), mainView)
		}
	} else {
		confirmDisable(app, iface.Name, mainView)
	}
}

func showInterfaceList(app *tview.Application, interfaces []net.Interface, mainView tview.Primitive) {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorYellow)

	list.SetTitle(" Select Ethernet Interface to Toggle ").
		SetTitleAlign(tview.AlignCenter).
		SetBorder(true)

	for _, iface := range interfaces {
		status, _ := getInterfaceStatus(iface.Name)
		list.AddItem(fmt.Sprintf("%s (%s)", iface.Name, status), "", 0, nil)
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			index := list.GetCurrentItem()
			if index < len(interfaces) {
				toggleInterface(app, interfaces[index].Name, mainView)
			} else {
				uiutil.IsFloatingBoxActive = false
				app.SetRoot(mainView, true)
				if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
					app.SetFocus(outputBox)
				}
			}
			return nil
		case tcell.KeyEscape:
			uiutil.IsFloatingBoxActive = false
			app.SetRoot(mainView, true)
			if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
				app.SetFocus(outputBox)
			}
			return nil
		}
		return event
	})

	list.AddItem("Cancel", "", 0, func() {
		uiutil.IsFloatingBoxActive = false
		app.SetRoot(mainView, true)
		if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
			app.SetFocus(outputBox)
		}
	})

	// Calculate dynamic size
	width := 40                   // Minimum width
	height := len(interfaces) + 4 // +4 for title, border, and Cancel option

	for _, iface := range interfaces {
		if len(iface.Name)+10 > width { // +10 for status and padding
			width = len(iface.Name) + 10
		}
	}

	// Create the floating box without showing it
	floatingBox := uiutil.FloatingBox(list, width, height)

	// Create a new page with the floating box
	page := tview.NewPages().
		AddPage("background", mainView, true, true).
		AddPage("floating", floatingBox, true, true)

	// Set the root of the application to the new page
	app.SetRoot(page, true).SetFocus(list)

	// Set IsFloatingBoxActive to true
	uiutil.IsFloatingBoxActive = true
}

func toggleInterface(app *tview.Application, name string, mainView tview.Primitive) {
	status, err := getInterfaceStatus(name)
	if err != nil {
		uiutil.ShowError(app, "Error getting interface status: "+err.Error(), mainView)
		return
	}

	newStatus := "up"
	if status == "up" {
		newStatus = "down"
	}

	err = setInterfaceStatus(name, newStatus)
	if err != nil {
		uiutil.ShowError(app, fmt.Sprintf("Error setting interface %s to %s: %s", name, newStatus, err.Error()), mainView)
	} else {
		uiutil.ShowMessage(app, fmt.Sprintf("Interface %s has been set to %s.", name, newStatus), mainView)
	}
}

func confirmDisable(app *tview.Application, name string, mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to disable interface %s?", name)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				err := setInterfaceStatus(name, "down")
				if err != nil {
					uiutil.ShowError(app, "Error disabling interface: "+err.Error(), mainView)
				} else {
					uiutil.ShowMessage(app, fmt.Sprintf("Interface %s has been disabled.", name), mainView)
				}
			} else {
				app.SetRoot(mainView, true)
				if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
					app.SetFocus(outputBox)
				}
			}
		})

	pages := tview.NewPages().
		AddPage("background", mainView, true, true).
		AddPage("modal", modal, true, true)

	app.SetRoot(pages, true).SetFocus(modal)
}

func getInterfaceStatus(name string) (string, error) {
	cmd := exec.Command("ip", "link", "show", name)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(string(output), "state UP") {
		return "up", nil
	}
	return "down", nil
}

func setInterfaceStatus(name string, status string) error {
	cmd := exec.Command("ip", "link", "set", name, status)
	return cmd.Run()
}
