package uiutil

import (
	"fmt"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PromptInput displays a modal to get input from the user.
// If prefill is not empty, the input field is prefilled with it.
func PromptInput(app *tview.Application, pages *tview.Pages, title, label string, mainView tview.Primitive, prefill ...string) (string, error) {
	input := tview.NewInputField().
		SetLabel(label).
		SetFieldWidth(40).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				app.QueueUpdateDraw(func() {
					pages.RemovePage("inputModal")
				})
			}
		})
	if len(prefill) > 0 {
		input.SetText(prefill[0])
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s] %s", title, label)).
		AddButtons([]string{"Submit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Submit" {
				text := input.GetText()
				app.QueueUpdateDraw(func() {
					pages.RemovePage("inputModal")
				})
				// Send the input text back
				app.SetFocus(mainView)
				// Store the input text somewhere accessible
				// For simplicity, using a global variable or a channel is recommended
				// Here, we'll just log it
				logger.Info("User input: %s", text)
			} else {
				app.QueueUpdateDraw(func() {
					pages.RemovePage("inputModal")
				})
			}
		})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(modal, 0, 1, false)

	pages.AddPage("inputModal", flex, true, true)

	// Start a goroutine to wait for user input
	var userInput string
	var errResult error
	done := make(chan struct{})

	go func() {
		app.Run()
		userInput = input.GetText()
		close(done)
	}()

	<-done

	if userInput == "" {
		errResult = fmt.Errorf("no input provided")
	}

	return userInput, errResult
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
