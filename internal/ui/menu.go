package ui

import (
	"github.com/fortifyde/netutil/internal/scripts"
	"github.com/rivo/tview"
)

func RunApp() error {
	app := tview.NewApplication()
	outputBox := tview.NewTextView().SetDynamicColors(true)
	outputBox.SetBorder(true).SetTitle("Script Output")

	menu := tview.NewList().
		AddItem("Run Script 1", "First Script", '1', func() {
			outputBox.Clear()
			scripts.RunBashScript("myscript1.sh", outputBox)
		}).
		AddItem("Run Script 2", "Second Script", '2', func() {
			outputBox.Clear()
			scripts.RunBashScript("myscript2.sh", outputBox)
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	flex := tview.NewFlex().
		AddItem(menu, 0, 1, true).
		AddItem(outputBox, 0, 2, false)

	return app.SetRoot(flex, true).Run()
}
