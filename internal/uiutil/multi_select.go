package uiutil

import (
	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowMultiSelect displays a list with multiple selection capability
func ShowMultiSelect(app *tview.Application, pages *tview.Pages,
	modalID, title string, items []string,
	onConfirm func([]string), mainView tview.Primitive) {

	modalManager.Enqueue(func(done func()) {
		selected := make(map[int]bool)
		list := tview.NewList().
			ShowSecondaryText(false).
			SetHighlightFullLine(true).
			SetSelectedTextColor(pkg.NordBg).
			SetSelectedBackgroundColor(pkg.NordHighlight).
			SetMainTextColor(pkg.NordFg)

		// Add items to the list
		for _, item := range items {
			list.AddItem(item, "", 0, nil)
		}

		// Create buttons flex container
		buttons := tview.NewFlex().
			SetDirection(tview.FlexColumn)

		// Create and configure buttons
		submitButton := tview.NewButton("Submit").
			SetSelectedFunc(func() {
				var selectedItems []string
				for i := range items {
					if selected[i] {
						selectedItems = append(selectedItems, items[i])
					}
				}
				CloseModalNotification(app, pages, modalID, mainView)
				onConfirm(selectedItems)
				done()
			})

		selectAllButton := tview.NewButton("Select All").
			SetSelectedFunc(func() {
				for i := range items {
					selected[i] = true
					list.SetItemText(i, "✓ "+items[i], "")
				}
			})

		cancelButton := tview.NewButton("Cancel").
			SetSelectedFunc(func() {
				CloseModalNotification(app, pages, modalID, mainView)
				done()
			})

		buttons.AddItem(submitButton, 0, 1, false).
			AddItem(tview.NewBox(), 1, 0, false).
			AddItem(selectAllButton, 0, 1, false).
			AddItem(tview.NewBox(), 1, 0, false).
			AddItem(cancelButton, 0, 1, false)

		// Create layout
		listFlex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(list, 0, 1, true).
			AddItem(buttons, 1, 0, false)

		listFlex.SetBorder(true).
			SetBorderColor(pkg.NordAccent)

		// Add title frame
		frame := tview.NewFrame(listFlex).
			SetBorders(0, 0, 0, 0, 0, 0).
			AddText(title, true, tview.AlignCenter, pkg.NordHighlight)

		// Create centered flex container
		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().
				AddItem(nil, 0, 1, false).
				AddItem(frame, 0, 3, true).
				AddItem(nil, 0, 1, false),
				0, 1, true).
			AddItem(nil, 0, 1, false)

		// Modify input capture to handle only space key for selection
		list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				CloseModalNotification(app, pages, modalID, mainView)
				done()
				return nil
			case tcell.KeyTab:
				if app.GetFocus() == list {
					app.SetFocus(submitButton)
				} else if app.GetFocus() == submitButton {
					app.SetFocus(selectAllButton)
				} else if app.GetFocus() == selectAllButton {
					app.SetFocus(cancelButton)
				} else {
					app.SetFocus(list)
				}
				return nil
			case tcell.KeyRune:
				if event.Rune() == ' ' {
					currentItem := list.GetCurrentItem()
					selected[currentItem] = !selected[currentItem]
					if selected[currentItem] {
						list.SetItemText(currentItem, "✓ "+items[currentItem], "")
					} else {
						list.SetItemText(currentItem, items[currentItem], "")
					}
				}
				return nil
			}
			return event
		})

		pages.AddPage(modalID, flex, true, true)
		app.SetFocus(list)
		SetFloatingBoxActive(true)
	})
}
