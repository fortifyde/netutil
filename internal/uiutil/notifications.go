package uiutil

import (
	"sync"

	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	modalManager *ModalManager
	once         sync.Once
)

// initializes the ModalManager singleton
func InitializeModalManager(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) {
	once.Do(func() {
		modalManager = NewModalManager(app, pages, mainView)
	})
}

var (
	IsFloatingBoxActive bool
	mutex               sync.Mutex
)

// safely set the flag
func SetFloatingBoxActive(active bool) {
	mutex.Lock()
	defer mutex.Unlock()
	IsFloatingBoxActive = active
}

// retrieves the flag safely
func GetFloatingBoxActive() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return IsFloatingBoxActive
}

// gracefully closes the notification modal window
func CloseModalNotification(app *tview.Application, pages *tview.Pages, modalName string, mainView tview.Primitive) {
	go app.QueueUpdateDraw(func() {
		pages.RemovePage(modalName)
		SetFloatingBoxActive(false)
		if toolbox, ok := mainView.(*tview.List); ok {
			app.SetFocus(toolbox)
		} else if mainView != nil {
			app.SetFocus(mainView)
		}
	})
}

// displays a simple message modal
func ShowMessage(app *tview.Application, pages *tview.Pages, modalName, message string, mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
		modal := tview.NewModal().
			SetText(message).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			})

		pages.AddPage(modalName, modal, true, true)
		app.SetFocus(modal)
		SetFloatingBoxActive(true)
	})
}

// displays an error modal with an optional callback
func ShowError(app *tview.Application, pages *tview.Pages, modalName, message string, mainView tview.Primitive, callback func()) {
	modalManager.Enqueue(func(done func()) {
		modal := tview.NewModal().
			SetText("[red]Error:[-] " + message).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				CloseModalNotification(app, pages, modalName, mainView)
				if callback != nil {
					// Enqueue the callback as a new modal operation.
					modalManager.Enqueue(func(doneCallback func()) {
						callback()
						doneCallback()
					})
				}
				done()
			})

		pages.AddPage(modalName, modal, true, true)
		app.SetFocus(modal)
		SetFloatingBoxActive(true)
	})
}

// displays a selectable list modal
func ShowList(app *tview.Application, pages *tview.Pages, modalName, title string, items []string, selectedFunc func(int), mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
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
				done()
			})
		}

		listFlex := tview.NewFlex().SetDirection(tview.FlexRow).
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
				done()
				return nil
			}
			return event
		})

		pages.AddPage(modalName, flex, true, true)
		app.SetFocus(list)
		SetFloatingBoxActive(true)
	})
}

// displays a confirmation modal with Yes/No options
func ShowConfirm(app *tview.Application, pages *tview.Pages, modalName, message string, callback func(bool), mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
		modal := tview.NewModal().
			SetText(message).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					callback(true)
				} else {
					callback(false)
				}
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			})

		pages.AddPage(modalName, modal, true, true)
		app.SetFocus(modal)
		SetFloatingBoxActive(true)
	})
}
