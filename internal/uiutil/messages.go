package uiutil

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CloseModal closes the modal and returns focus to the output box
func CloseModal(app *tview.Application, mainView tview.Primitive) {
	IsFloatingBoxActive = false
	app.SetRoot(mainView, true)
	if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
		app.SetFocus(outputBox)
	}
}

// ShowMessage displays a message modal over the main view
func ShowMessage(app *tview.Application, message string, mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModal(app, mainView)
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

// CloseMessage closes the message modal and returns focus to the output box
func CloseMessage(app *tview.Application, mainView tview.Primitive) {
	IsFloatingBoxActive = false
	app.SetRoot(mainView, true)
	if outputBox, ok := mainView.(*tview.Flex).GetItem(1).(*tview.Flex).GetItem(1).(*tview.List); ok {
		app.SetFocus(outputBox)
	}
}

// SetupFloatingBoxInputCapture sets up input capture for a floating box
func SetupFloatingBoxInputCapture(modal *tview.Modal, app *tview.Application, mainView tview.Primitive, customAction func()) {
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if customAction != nil {
				customAction()
			} else {
				CloseMessage(app, mainView)
			}
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				CloseMessage(app, mainView)
				return nil
			}
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return nil // Return nil to prevent any other input handling
	})
}

// ShowTimedMessage displays a message modal over the main view and closes it after a specified duration
func ShowTimedMessage(app *tview.Application, message string, mainView tview.Primitive, duration time.Duration) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModal(app, mainView)
		})

	pages := tview.NewPages().
		AddPage("background", mainView, true, true).
		AddPage("modal", modal, true, true)

	app.SetRoot(pages, true).SetFocus(modal)

	// Close the modal automatically after the specified duration
	go func() {
		time.Sleep(duration)
		app.QueueUpdateDraw(func() {
			CloseModal(app, mainView)
		})
	}()
}
