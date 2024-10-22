package ui

import (
	"fmt"
	"time"

	"github.com/fortifyde/netutil/internal/functions"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

func RunApp() error {
	logger.Info("Starting UI application")
	app := tview.NewApplication()
	pages := tview.NewPages()

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
	categories := []string{"System Configuration", "Category 2", "Category 3", "Category 4", "Category 5"}
	categoryContents := make(map[string][]string)

	categoryContents["System Configuration"] = []string{"Check and toggle interfaces", "Edit Working Directory", "Save Network Config", "Load Network Config"}

	for _, category := range categories {
		menu.AddItem(category, "", 0, nil)
		if category != "System Configuration" {
			functions := []string{}
			for j := 1; j <= 3; j++ {
				functions = append(functions, fmt.Sprintf("Function %d", j))
			}
			categoryContents[category] = functions
		}
	}

	updatetoolbox := func(category string) {
		toolbox.Clear()
		for _, function := range categoryContents[category] {
			toolbox.AddItem(function, "", 0, func() {
				switch function {
				case "Check and toggle interfaces":
					functions.ToggleEthernetInterfaces(app, pages, toolbox)
				case "Edit Working Directory":
					if err := functions.EditWorkingDirectory(app, pages, toolbox); err != nil {
						uiutil.ShowMessage(app, pages, fmt.Sprintf("Error: %s", err), toolbox)
					}
				case "Save Network Config":
					if err := functions.SaveNetworkConfig(app, pages, toolbox); err != nil {
						uiutil.ShowMessage(app, pages, fmt.Sprintf("Error: %s", err), toolbox)
					}
				case "Load Network Config":
					if err := functions.LoadAndApplyNetworkConfig(app, pages, toolbox); err != nil {
						uiutil.ShowMessage(app, pages, fmt.Sprintf("Error: %s", err), toolbox)
					}
				default:
					uiutil.ShowMessage(app, pages, fmt.Sprintf("Function '%s' not implemented yet", function), toolbox)
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
