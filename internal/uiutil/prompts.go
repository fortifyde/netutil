package uiutil

import (
	"fmt"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/rivo/tview"
)

// gracefully closes the Input Prompt Modal window
func CloseInputModal(app *tview.Application, pages *tview.Pages, modalName string, mainView tview.Primitive) {
	pages.RemovePage(modalName)
	SetFloatingBoxActive(false)
	if mainView != nil {
		app.SetFocus(mainView)
	}
}

// gracefully closes the Confirmation Prompt Modal window
func CloseConfirmModal(app *tview.Application, pages *tview.Pages, modalName string, mainView tview.Primitive) {
	go app.QueueUpdateDraw(func() {
		pages.RemovePage(modalName)
		SetFloatingBoxActive(false)
		if mainView != nil {
			app.SetFocus(mainView)
		}
	})
}

// displays a modal to get input from the user
func PromptInput(app *tview.Application, pages *tview.Pages, modalName, title, label string, mainView tview.Primitive, callback func(string, error), prefill ...string) {
	modalManager.Enqueue(func(done func()) {
		input := tview.NewInputField().
			SetLabel(label).
			SetFieldWidth(40)

		if len(prefill) > 0 {
			input.SetText(prefill[0])
		}

		form := tview.NewForm().
			AddFormItem(input).
			AddButton("Submit", func() {
				text := strings.TrimSpace(input.GetText())
				if text == "" && strings.Contains(modalName, "dirName") {
					ShowError(app, pages, modalName+"_error",
						fmt.Sprintf("%s cannot be empty. Please enter a valid name.", label),
						mainView,
						func() {
							PromptInput(app, pages, modalName, title, label, mainView, callback, prefill...)
						})
				} else {
					logger.Info("User input for %s: %s", label, text)
					callback(text, nil)
					CloseModalNotification(app, pages, modalName, mainView)
					done()
				}
			}).
			AddButton("Cancel", func() {
				callback("", fmt.Errorf("input canceled"))
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			})

		form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)

		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 8, 1, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage(modalName, flex, true, true)

		app.SetFocus(input)
		SetFloatingBoxActive(true)
	})
}

// displays a confirmation modal to the user
func PromptConfirmation(app *tview.Application, pages *tview.Pages, modalName, title, message string, callback func(bool, error), mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("[%s] %s", title, message)).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					callback(true, nil)
				} else {
					callback(false, fmt.Errorf("user declined"))
				}
				CloseConfirmModal(app, pages, modalName, mainView)
				done()
			})

		modal.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)

		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modal, 4, 1, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage(modalName, flex, true, true)
		app.SetFocus(modal)
		SetFloatingBoxActive(true)
	})
}
