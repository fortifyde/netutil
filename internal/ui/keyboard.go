package ui

import (
	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// set up main keyboard controls for application
func SetupKeyboardControls(app *tview.Application, menu *tview.List, toolbox *tview.List, pages *tview.Pages) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if uiutil.GetFloatingBoxActive() {
			return event
		}

		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyCtrlC:
			return nil
		case tcell.KeyLeft, tcell.KeyRight:
			return handleHorizontalNavigation(app, menu, toolbox, event)
		case tcell.KeyUp, tcell.KeyDown:
			return handleVerticalNavigation(app, menu, toolbox, event)
		case tcell.KeyTab:
			return handleTabNavigation(menu)
		}

		switch event.Rune() {
		case 'h':
			return handleHorizontalNavigation(app, menu, toolbox, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone))
		case 'l':
			return handleHorizontalNavigation(app, menu, toolbox, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
		case 'k':
			return handleVerticalNavigation(app, menu, toolbox, tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
		case 'j':
			return handleVerticalNavigation(app, menu, toolbox, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
		case 'q':
			app.Stop()
			return nil
		case 'd':
			return nil
		}

		return event
	})
}

// manages left/right navigation between menu and command box
func handleHorizontalNavigation(app *tview.Application, menu *tview.List, toolbox *tview.List, event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyLeft {
		app.SetFocus(menu)
		menu.SetSelectedTextColor(pkg.NordBg)
		menu.SetSelectedBackgroundColor(pkg.NordHighlight)
		toolbox.SetSelectedTextColor(pkg.NordFg)
		toolbox.SetSelectedBackgroundColor(pkg.NordBg)
	} else if event.Key() == tcell.KeyRight {
		app.SetFocus(toolbox)
		menu.SetSelectedTextColor(pkg.NordAccent)
		menu.SetSelectedBackgroundColor(pkg.NordBg)
		toolbox.SetSelectedTextColor(pkg.NordBg)
		toolbox.SetSelectedBackgroundColor(pkg.NordHighlight)
	}
	return nil
}

// manages up/down navigation within menu or command box
func handleVerticalNavigation(app *tview.Application, menu *tview.List, toolbox *tview.List, event *tcell.EventKey) *tcell.EventKey {
	focused := app.GetFocus()
	if focused == menu {
		if event.Key() == tcell.KeyUp {
			menu.SetCurrentItem(menu.GetCurrentItem() - 1)
		} else {
			menu.SetCurrentItem(menu.GetCurrentItem() + 1)
		}
	} else if focused == toolbox {
		if event.Key() == tcell.KeyUp {
			toolbox.SetCurrentItem(toolbox.GetCurrentItem() - 1)
		} else {
			toolbox.SetCurrentItem(toolbox.GetCurrentItem() + 1)
		}
	}
	return nil
}

// cycles through menu items
func handleTabNavigation(menu *tview.List) *tcell.EventKey {
	currentItem := menu.GetCurrentItem()
	itemCount := menu.GetItemCount()
	menu.SetCurrentItem((currentItem + 1) % itemCount)
	return nil
}
