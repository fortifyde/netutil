package ui

import (
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func SetupKeyboardControls(app *tview.Application, menu *tview.List, outputBox *tview.List) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if uiutil.IsFloatingBoxActive {
			// We're in a floating box, let it handle its own input
			return event
		}

		// Add this block to handle the focus return to outputBox
		if event.Key() == tcell.KeyRune && event.Rune() == 0 {
			app.SetFocus(outputBox)
			menu.SetSelectedTextColor(nordAccent)
			menu.SetSelectedBackgroundColor(nordBg)
			outputBox.SetSelectedTextColor(nordBg)
			outputBox.SetSelectedBackgroundColor(nordHighlight)
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyLeft:
			app.SetFocus(menu)
			menu.SetSelectedTextColor(nordBg)
			menu.SetSelectedBackgroundColor(nordHighlight)
			outputBox.SetSelectedTextColor(nordFg)
			outputBox.SetSelectedBackgroundColor(nordBg)
			return nil
		case tcell.KeyRight:
			app.SetFocus(outputBox)
			menu.SetSelectedTextColor(nordAccent)
			menu.SetSelectedBackgroundColor(nordBg)
			outputBox.SetSelectedTextColor(nordBg)
			outputBox.SetSelectedBackgroundColor(nordHighlight)
			return nil
		case tcell.KeyUp:
			if app.GetFocus() == menu {
				menu.SetCurrentItem(menu.GetCurrentItem() - 1)
			} else if app.GetFocus() == outputBox {
				outputBox.SetCurrentItem(outputBox.GetCurrentItem() - 1)
			}
			return nil
		case tcell.KeyDown:
			if app.GetFocus() == menu {
				menu.SetCurrentItem(menu.GetCurrentItem() + 1)
			} else if app.GetFocus() == outputBox {
				outputBox.SetCurrentItem(outputBox.GetCurrentItem() + 1)
			}
			return nil
		}

		// Handle rune inputs separately
		switch event.Rune() {
		case 'h':
			app.SetFocus(menu)
			menu.SetSelectedTextColor(nordBg)
			menu.SetSelectedBackgroundColor(nordHighlight)
			outputBox.SetSelectedTextColor(nordFg)
			outputBox.SetSelectedBackgroundColor(nordBg)
			return nil
		case 'l':
			app.SetFocus(outputBox)
			menu.SetSelectedTextColor(nordAccent)
			menu.SetSelectedBackgroundColor(nordBg)
			outputBox.SetSelectedTextColor(nordBg)
			outputBox.SetSelectedBackgroundColor(nordHighlight)
			return nil
		case 'k':
			if app.GetFocus() == menu {
				menu.SetCurrentItem(menu.GetCurrentItem() - 1)
			} else if app.GetFocus() == outputBox {
				outputBox.SetCurrentItem(outputBox.GetCurrentItem() - 1)
			}
			return nil
		case 'j':
			if app.GetFocus() == menu {
				menu.SetCurrentItem(menu.GetCurrentItem() + 1)
			} else if app.GetFocus() == outputBox {
				outputBox.SetCurrentItem(outputBox.GetCurrentItem() + 1)
			}
			return nil
		case 'q':
			app.Stop()
			return nil
		case 'd':
			// TODO: Implement function description
			return nil
		}
		return event
	})
}
