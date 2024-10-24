package uiutil

import (
	"fmt"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/rivo/tview"
)

// CloseInputModal gracefully closes the Input Prompt Modal window.
func CloseInputModal(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) {
	IsFloatingBoxActive = false
	pages.RemovePage("inputModal")
	if mainView != nil {
		app.SetFocus(mainView)
	}
}

// CloseConfirmModal gracefully closes the Confirmation Prompt Modal window.
func CloseConfirmModal(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) {
	IsFloatingBoxActive = false
	pages.RemovePage("confirmModal")
	if mainView != nil {
		app.SetFocus(mainView)
	}
}

// PromptInput displays a modal to get input from the user.
func PromptInput(app *tview.Application, pages *tview.Pages, title, label string, mainView tview.Primitive, callback func(string, error), prefill ...string) {
	input := tview.NewInputField().
		SetLabel(label).
		SetFieldWidth(40)

	if len(prefill) > 0 {
		input.SetText(prefill[0])
	}

	form := tview.NewForm().
		AddFormItem(input).
		AddButton("Submit", func() {
			text := input.GetText()
			logger.Info("User input: %s", text)
			callback(text, nil)
			CloseInputModal(app, pages, mainView)
		}).
		AddButton("Cancel", func() {
			callback("", fmt.Errorf("input canceled"))
			CloseInputModal(app, pages, mainView)
		})

	form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(form, 8, 1, true). // Adjusted height
		AddItem(nil, 0, 1, false)

	// Open the modal
	pages.AddPage("inputModal", flex, true, true)
	app.SetFocus(input)
	IsFloatingBoxActive = true
}

// PromptConfirmation displays a confirmation modal to the user.
func PromptConfirmation(app *tview.Application, pages *tview.Pages, title, message string, callback func(bool, error), mainView tview.Primitive) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s] %s", title, message)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				callback(true, nil)
			} else {
				callback(false, fmt.Errorf("user declined"))
			}
			CloseConfirmModal(app, pages, mainView)
		})

	modal.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 4, 1, true). // Adjusted height
		AddItem(nil, 0, 1, false)

	// Open the modal
	pages.AddPage("confirmModal", flex, true, true)
	app.SetFocus(modal)
	IsFloatingBoxActive = true
}
