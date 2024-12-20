package ui

import (
	"fmt"
	"time"

	"github.com/fortifyde/netutil/internal/functions"
	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

func RunApp(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) error {
	logger.Info("Starting UI application")

	// set default colors
	tview.Styles.PrimitiveBackgroundColor = pkg.NordBg
	tview.Styles.ContrastBackgroundColor = pkg.NordBg
	tview.Styles.MoreContrastBackgroundColor = pkg.NordBg
	tview.Styles.PrimaryTextColor = pkg.NordFg
	tview.Styles.SecondaryTextColor = pkg.NordFg
	tview.Styles.TertiaryTextColor = pkg.NordFg
	tview.Styles.InverseTextColor = pkg.NordBg
	tview.Styles.ContrastSecondaryTextColor = pkg.NordHighlight

	titleBox := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[::b]NetUtil[-] [::i]v%s[-]", pkg.Version)).
		SetTextColor(pkg.NordAccent).
		SetDynamicColors(true)
	titleBox.SetBorder(true)

	searchBox := tview.NewInputField().
		SetLabel("Press / to start search").
		SetFieldTextColor(pkg.NordFg).
		SetFieldWidth(0)
	searchBox.SetBorder(true).SetTitle("Search").SetTitleAlign(tview.AlignLeft)

	lastCompileDate := time.Now().Format("2006-01-02")
	toolbox := tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(pkg.NordFg).
		SetSelectedTextColor(pkg.NordFg).
		SetSelectedBackgroundColor(pkg.NordBg)
	toolbox.SetBorder(true).SetTitle("Network Toolbox - " + lastCompileDate).SetTitleAlign(tview.AlignLeft)

	menu := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(pkg.NordBg).
		SetSelectedBackgroundColor(pkg.NordHighlight).
		SetMainTextColor(pkg.NordFg)
	menu.SetBorder(true).SetTitle("Main Menu").SetTitleAlign(tview.AlignLeft)

	categories := []string{"System Configuration", "Network Recon", "Category 3", "Category 4", "Category 5"}
	categoryContents := make(map[string][]string)

	categoryContents["System Configuration"] = []string{
		"Check and toggle interfaces",
		"Configure Network Interfaces",
		"Edit Working Directory",
		"Save Network Config",
		"Load Network Config",
	}
	categoryContents["Network Recon"] = []string{"Wireshark Listening", "Discovery Scan"} // Added "Discovery Scan"

	for _, category := range categories {
		menu.AddItem(category, "", 0, nil)
		if _, exists := categoryContents[category]; !exists {
			categoryContents[category] = []string{"Function 1", "Function 2", "Function 3"}
		}
	}

	updatetoolbox := func(category string) {
		toolbox.Clear()
		for _, function := range categoryContents[category] {
			function := function
			toolbox.AddItem(function, "", 0, func() {
				switch category {
				case "System Configuration":
					switch function {
					case "Check and toggle interfaces":
						if err := functions.ToggleEthernetInterfaces(app, pages, toolbox); err != nil {
							uiutil.ShowError(app, pages, "toggleInterfacesErrorModal", fmt.Sprintf("Error: %s", err), toolbox, nil)
						}
					case "Configure Network Interfaces":
						if err := functions.ConfigureNetworkInterfaces(app, pages, toolbox); err != nil {
							uiutil.ShowError(app, pages, "configureNetworkInterfacesErrorModal", fmt.Sprintf("Error: %s", err), toolbox, nil)
						}
					case "Edit Working Directory":
						if err := configuration.EditWorkingDirectory(app, pages, toolbox); err != nil {
							uiutil.ShowError(app, pages, "editWorkingDirectoryErrorModal", fmt.Sprintf("Error: %s", err), toolbox, nil)
						}
					case "Save Network Config":
						if err := functions.SaveNetworkConfig(app, pages, toolbox); err != nil {
							uiutil.ShowError(app, pages, "saveNetworkConfigErrorModal", fmt.Sprintf("Error: %s", err), toolbox, nil)
						}
					case "Load Network Config":
						if err := functions.LoadAndApplyNetworkConfig(app, pages, toolbox); err != nil {
							uiutil.ShowError(app, pages, "loadNetworkConfigErrorModal", fmt.Sprintf("Error: %s", err), toolbox, nil)
						}
					}
				case "Network Recon":
					switch function {
					case "Wireshark Listening":
						err := functions.StartWiresharkListening(app, pages, toolbox)
						if err != nil {
							uiutil.ShowError(app, pages, "wiresharkListeningErrorModal", fmt.Sprintf("Wireshark listening error: %v", err), toolbox, nil)
						}
					case "Discovery Scan":
						err := functions.StartDiscoveryScan(app, pages, toolbox)
						if err != nil {
							uiutil.ShowError(app, pages, "discoveryScanErrorModal", fmt.Sprintf("Discovery Scan error: %v", err), toolbox, nil)
						}
					}
				default:
					uiutil.ShowMessage(app, pages, "functionNotImplementedModal", fmt.Sprintf("Function '%s' not implemented yet", function), toolbox)
				}
			})
		}
	}

	menu.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		updatetoolbox(mainText)
	})

	// populate toolbox initially
	updatetoolbox(categories[0])

	cmdInfoView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(createCommandInfoText())

	cmdInfoView.SetBorder(true).
		SetTitle("Command List").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(pkg.NordAccent)

	// adjust height based on lines in cmdbox
	cmdInfoHeight := 7

	menuWidth := 30

	grid := tview.NewGrid().
		SetRows(3, 0, cmdInfoHeight).
		SetColumns(menuWidth, 0).
		SetBorders(false).
		AddItem(titleBox, 0, 0, 1, 1, 0, 0, false).
		AddItem(searchBox, 0, 1, 1, 1, 0, 0, false).
		AddItem(menu, 1, 0, 1, 1, 0, 0, true).
		AddItem(toolbox, 1, 1, 1, 1, 0, 0, false).
		AddItem(cmdInfoView, 2, 0, 1, 2, 0, 0, false)

	pages.AddPage("main", grid, true, true)

	SetupKeyboardControls(app, menu, toolbox, pages)

	logger.Info("UI application started successfully")
	return app.SetRoot(pages, true).Run()
}

func createCommandInfoText() string {
	return `[::b]Movement[::-]                    [::b]Other[::-]
h, ←: Focus menu              q, Ctrl+C: Exit NetUtil
l, →: Focus functions         d: Function description
k, ↑: Select item above       /: Search
j, ↓: Select item below`
}
