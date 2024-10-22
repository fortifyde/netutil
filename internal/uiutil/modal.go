package uiutil

import (
	"time"

	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var IsFloatingBoxActive bool

func CloseModal(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) {
	IsFloatingBoxActive = false
	pages.RemovePage("modal")
	if toolbox, ok := mainView.(*tview.List); ok {
		app.SetFocus(toolbox)
	}
}

func ShowMessage(app *tview.Application, pages *tview.Pages, message string, mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModal(app, pages, mainView)
		})

	IsFloatingBoxActive = true
	pages.AddPage("modal", modal, true, true)
	app.SetFocus(modal)
}

func ShowError(app *tview.Application, pages *tview.Pages, message string, mainView tview.Primitive) {
	ShowMessage(app, pages, "Error: "+message, mainView)
}

func ShowTimedMessage(app *tview.Application, pages *tview.Pages, message string, mainView tview.Primitive, duration time.Duration) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModal(app, pages, mainView)
		})

	IsFloatingBoxActive = true
	pages.AddPage("modal", modal, true, true)
	app.SetFocus(modal)

	go func() {
		time.Sleep(duration)
		app.QueueUpdateDraw(func() {
			if pages.HasPage("modal") {
				CloseModal(app, pages, mainView)
			}
		})
	}()
}

func ShowList(app *tview.Application, pages *tview.Pages, title string, items []string, selectedFunc func(int), mainView tview.Primitive) {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(pkg.NordBg).
		SetSelectedBackgroundColor(pkg.NordHighlight).
		SetMainTextColor(pkg.NordFg)

	for i, item := range items {
		index := i
		list.AddItem(item, "", 0, func() {
			CloseModal(app, pages, mainView)
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
			CloseModal(app, pages, mainView)
			return nil
		}
		return event
	})

	IsFloatingBoxActive = true
	pages.AddPage("modal", flex, true, true)
	app.SetFocus(list)
}

func ShowConfirm(app *tview.Application, pages *tview.Pages, message string, callback func(bool), mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			CloseModal(app, pages, mainView)
			callback(buttonLabel == "Yes")
		})

	IsFloatingBoxActive = true
	pages.AddPage("modal", modal, true, true)
	app.SetFocus(modal)
}
