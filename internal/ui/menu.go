package ui

import (
	"fmt"
	"time"

	"github.com/fortifyde/netutil/internal/functions"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Nord color palette
var (
	nordBg        = tcell.NewRGBColor(46, 52, 64)    // Nord0
	nordFg        = tcell.NewRGBColor(216, 222, 233) // Nord4
	nordHighlight = tcell.NewRGBColor(235, 203, 139) // Nord13
	nordAccent    = tcell.NewRGBColor(191, 97, 106)  // Nord12
)

func RunApp() error {
	app := tview.NewApplication()
	var mainFlex *tview.Flex

	// Set default colors
	tview.Styles.PrimitiveBackgroundColor = nordBg
	tview.Styles.ContrastBackgroundColor = nordBg
	tview.Styles.MoreContrastBackgroundColor = nordBg
	tview.Styles.PrimaryTextColor = nordFg
	tview.Styles.SecondaryTextColor = nordFg
	tview.Styles.TertiaryTextColor = nordFg
	tview.Styles.InverseTextColor = nordBg
	tview.Styles.ContrastSecondaryTextColor = nordHighlight

	// Title and version box
	titleBox := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("[::b]NetUtil[-] [::i]v1.0.0[-]").
		SetTextColor(nordAccent).
		SetDynamicColors(true)
	titleBox.SetBorder(true)

	// Search box
	searchBox := tview.NewInputField().
		SetLabel("Search: ").
		SetFieldTextColor(nordFg).
		SetFieldWidth(30).
		SetTitleAlign(tview.AlignLeft)
	searchBox.SetBorder(true).SetTitle("Search")

	// Output box
	lastCompileDate := time.Now().Format("2006-01-02")
	outputBox := tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(nordFg).
		SetSelectedTextColor(nordFg).
		SetSelectedBackgroundColor(nordBg)
	outputBox.SetBorder(true).SetTitle("Network Toolbox - " + lastCompileDate).SetTitleAlign(tview.AlignLeft)

	// Menu
	menu := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(nordBg).
		SetSelectedBackgroundColor(nordHighlight).
		SetMainTextColor(nordFg)

	categories := []string{"System Configuration", "Category 2", "Category 3", "Category 4", "Category 5"}
	categoryContents := make(map[string][]string)

	// Define functions for the "System Configuration" category
	categoryContents["System Configuration"] = []string{"Check and toggle interfaces", "Function 2", "Function 3"}

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

	updateOutputBox := func(category string) {
		outputBox.Clear()
		for _, function := range categoryContents[category] {
			outputBox.AddItem(function, "", 0, func() {
				if function == "Check and toggle interfaces" {
					functions.ToggleEthernetInterfaces(app, mainFlex)
				} else {
					showMessage(app, fmt.Sprintf("Function '%s' not implemented yet", function))
				}
			})
		}
	}

	menu.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		updateOutputBox(mainText)
	})

	menu.SetBorder(true).SetTitle("Menu")

	// Populate outputBox with initial category content
	updateOutputBox(categories[0])

	// Command info box
	cmdInfoBox := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf(""+
			"[%s]Movement:[-]                         [%s]Other:[-]\n"+
			"[%s]h, ←[-]: Focus menu                  [%s]q, Ctrl+C[-]: Exit NetUtil\n"+
			"[%s]l, →[-]: Focus functions             [%s]d[-]        : Function description\n"+
			"[%s]k, ↑[-]: Select item above           [%s]/[-]        : Search\n"+
			"[%s]j, ↓[-]: Select item below",
			nordHighlight.String(),
			nordHighlight.String(),
			nordAccent.String(),
			nordAccent.String(),
			nordAccent.String(),
			nordAccent.String(),
			nordAccent.String(),
			nordAccent.String(),
			nordAccent.String()))
	cmdInfoBox.SetBorder(true).SetTitle("Command List").SetTitleAlign(tview.AlignLeft)

	// Layout
	topFlex := tview.NewFlex().
		AddItem(titleBox, 0, 1, false).
		AddItem(searchBox, 0, 2, false)

	middleFlex := tview.NewFlex().
		AddItem(menu, 0, 1, true).
		AddItem(outputBox, 0, 2, false)

	bottomFlex := tview.NewFlex().
		AddItem(cmdInfoBox, 0, 1, false)

	mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 3, 1, false).
		AddItem(middleFlex, 0, 1, true).
		AddItem(bottomFlex, 7, 1, false)

	SetupKeyboardControls(app, menu, outputBox)

	return app.SetRoot(mainFlex, true).Run()
}

// Add this helper function at the end of the file
func showMessage(app *tview.Application, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(app.GetFocus(), true)
		})

	app.SetRoot(modal, false).SetFocus(modal)
}
