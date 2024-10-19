package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func SetupKeyboardControls(app *tview.Application, menu *tview.List, outputBox *tview.List) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
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
		case 'd':
			// TODO: Implement function description
			return nil
		}
		return event
	})
}
