package uiutil

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var IsFloatingBoxActive bool

// FloatingBox creates a centered, floating box with dynamic size
func FloatingBox(content tview.Primitive, width, height int) *tview.Flex {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(content, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

// ShowFloatingBox displays a centered, floating box over the current application view
func ShowFloatingBox(app *tview.Application, content tview.Primitive, width, height int, mainView tview.Primitive, keepMainViewVisible bool) *tview.Pages {
	IsFloatingBoxActive = true
	floatingBox := FloatingBox(content, width, height)
	page := tview.NewPages().
		AddPage("background", mainView, true, keepMainViewVisible).
		AddPage("floating", floatingBox, true, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			CloseFloatingBox(app, mainView)
			return nil
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
			return event
		}

		// Handle vim-style navigation
		switch event.Rune() {
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'h':
			return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
		case 'q':
			CloseFloatingBox(app, mainView)
			return nil
		}

		// Pass other key events to the content
		if handler := content.InputHandler(); handler != nil {
			handler(event, func(p tview.Primitive) {
				// Do nothing, but satisfy the function signature
			})
		}
		return event
	})

	return page
}

// CloseFloatingBox closes the floating box and returns focus to the output box
func CloseFloatingBox(app *tview.Application, mainView tview.Primitive) {
	IsFloatingBoxActive = false
	app.SetRoot(mainView, true)
	if flex, ok := mainView.(*tview.Flex); ok {
		if middleFlex, ok := flex.GetItem(1).(*tview.Flex); ok {
			if outputBox, ok := middleFlex.GetItem(1).(*tview.List); ok {
				app.SetFocus(outputBox)
			}
		}
	}
}
