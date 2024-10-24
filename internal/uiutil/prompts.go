package uiutil

import (
	"fmt"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/rivo/tview"
)

// PromptInput displays a modal to get input from the user.
// If prefill is not empty, the input field is prefilled with it.
// The result is returned via the callback function.
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
			pages.RemovePage("inputModal")
			app.SetFocus(mainView)
		}).
		AddButton("Cancel", func() {
			callback("", fmt.Errorf("input cancelled"))
			pages.RemovePage("inputModal")
			app.SetFocus(mainView)
		})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true)

	pages.AddPage("inputModal", flex, true, true)
	app.SetFocus(input)
}

// PromptConfirmation displays a confirmation modal to the user.
func PromptConfirmation(app *tview.Application, pages *tview.Pages, title, message string) (bool, error) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s] %s", title, message)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				// Send confirmation back
			} else {
				// Send cancellation back
			}
		})

	pages.AddPage("confirmModal", modal, true, true)

	// Implement logic to capture user response
	// This requires a more complex setup, potentially using channels

	// Placeholder return
	return true, nil
}
