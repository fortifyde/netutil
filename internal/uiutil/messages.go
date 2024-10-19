package uiutil

import (
	"github.com/rivo/tview"
)

// ShowMessage displays a message modal over the main view
func ShowMessage(app *tview.Application, message string, mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(mainView, true)
			if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
				app.SetFocus(outputBox)
			}
		})

	pages := tview.NewPages().
		AddPage("background", mainView, true, true).
		AddPage("modal", modal, true, true)

	app.SetRoot(pages, true).SetFocus(modal)
}

// ShowError displays an error message modal over the main view
func ShowError(app *tview.Application, message string, mainView tview.Primitive) {
	ShowMessage(app, "Error: "+message, mainView)
}
