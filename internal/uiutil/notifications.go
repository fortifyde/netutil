package uiutil

import (
	"sync"

	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	IsFloatingBoxActive bool
	mutex               sync.Mutex
)

// Safely set the flag
func SetFloatingBoxActive(active bool) {
	mutex.Lock()
	defer mutex.Unlock()
	IsFloatingBoxActive = active
}

// Safely get the flag
func GetFloatingBoxActive() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return IsFloatingBoxActive
}

// CloseModal gracefully closes the notification modal window.
func CloseModalNotification(app *tview.Application, pages *tview.Pages, modalName string, mainView tview.Primitive) {
	pages.RemovePage(modalName)
	SetFloatingBoxActive(false)
	if toolbox, ok := mainView.(*tview.List); ok {
		app.SetFocus(toolbox)
	} else if mainView != nil {
		app.SetFocus(mainView)
	}
}

// ShowMessage displays a simple message modal.
func ShowMessage(app *tview.Application, pages *tview.Pages, modalName, message string, mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModalNotification(app, pages, modalName, mainView)
		})

	pages.AddPage(modalName, modal, true, true)
	app.SetFocus(modal)
	SetFloatingBoxActive(true)
}

// ShowError displays an error modal with an optional callback.
// modalName should be unique for each invocation.
func ShowError(app *tview.Application, pages *tview.Pages, modalName, message string, mainView tview.Primitive, callback func()) chan struct{} {
	done := make(chan struct{})
	modal := tview.NewModal().
		SetText("[red]Error:[-] " + message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModalNotification(app, pages, modalName, mainView)
			close(done)
			if callback != nil {
				go callback()
			}
		})

	pages.AddPage(modalName, modal, true, true)
	app.SetFocus(modal)
	SetFloatingBoxActive(true)
	return done
}

// ShowList displays a selectable list modal.
func ShowList(app *tview.Application, pages *tview.Pages, modalName, title string, items []string, selectedFunc func(int), mainView tview.Primitive) {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(pkg.NordBg).
		SetSelectedBackgroundColor(pkg.NordHighlight).
		SetMainTextColor(pkg.NordFg)

	for i, item := range items {
		index := i
		list.AddItem(item, "", 0, func() {
			CloseModalNotification(app, pages, modalName, mainView)
			selectedFunc(index)
		})
	}

	listFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true)

	listFlex.SetBorder(true).
		SetBorderColor(pkg.NordAccent)

	frame := tview.NewFrame(listFlex).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText(title, true, tview.AlignCenter, pkg.NordHighlight)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(frame, 0, 3, true).
			AddItem(nil, 0, 1, false),
			0, 1, true).
		AddItem(nil, 0, 1, false)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			CloseModalNotification(app, pages, modalName, mainView)
			return nil
		}
		return event
	})

	pages.AddPage(modalName, flex, true, true)
	app.SetFocus(list)
	SetFloatingBoxActive(true)
}

// ShowConfirm displays a confirmation modal with Yes/No options.
func ShowConfirm(app *tview.Application, pages *tview.Pages, modalName, message string, callback func(bool), mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModalNotification(app, pages, modalName, mainView)
			callback(buttonLabel == "Yes")
		})

	pages.AddPage(modalName, modal, true, true)
	app.SetFocus(modal)
	SetFloatingBoxActive(true)
}
