package uiutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/fortifyde/netutil/internal/logger"
	"github.com/rivo/tview"
	"golang.org/x/term"
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
		modalName = "input_" + modalName
		// Get terminal dimensions
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			// Fallback to reasonable defaults if we can't get terminal size
			width, height = 100, 30
		}

		// Calculate modal size (70% width)
		modalWidth := int(float64(width) * 0.7)

		input := tview.NewInputField().
			SetLabel(label).
			SetFieldWidth(modalWidth - len(label) - 4) // Subtract label length and some padding

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
				logger.Info("User canceled input")
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			})

		// Center align the buttons
		form.SetButtonsAlign(tview.AlignCenter)
		form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)

		// Calculate position for centering
		modalX := (width - modalWidth) / 2
		modalY := (height - 7) / 2 // 7 is approximate height of the form

		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, modalY, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, modalX, 1, false).
				AddItem(form, modalWidth, 1, true).
				AddItem(nil, modalX, 1, false),
				7, 1, true).
			AddItem(nil, modalY, 1, false)

		pages.AddPage(modalName, flex, true, true)

		app.SetFocus(input)
		SetFloatingBoxActive(true)
	})
}

// displays a confirmation modal to the user
func PromptConfirmation(app *tview.Application, pages *tview.Pages, modalName, title, message string, callback func(bool, error), mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
		modalName = "confirm_" + modalName
		modal := tview.NewModal().
			SetText(fmt.Sprintf("[%s] %s", title, message)).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					callback(true, nil)
				} else {
					logger.Info("User declined confirmation")
					callback(false, nil)
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

// displays a modal to get input from the user with an option to cancel all
func PromptInputWithCancelAll(app *tview.Application, pages *tview.Pages, modalName, title, label string, mainView tview.Primitive, callback func(string, error), cancelAllCallback func(), prefill ...string) {
	modalManager.Enqueue(func(done func()) {
		modalName = "input_" + modalName
		// Get terminal dimensions
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			// Fallback to reasonable defaults if we can't get terminal size
			width, height = 100, 30
		}

		// Calculate modal size (70% width)
		modalWidth := int(float64(width) * 0.7)

		input := tview.NewInputField().
			SetLabel(label).
			SetFieldWidth(modalWidth - len(label) - 4) // Subtract label length and some padding

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
							PromptInputWithCancelAll(app, pages, modalName, title, label, mainView, callback, cancelAllCallback, prefill...)
						})
				} else {
					logger.Info("User input for %s: %s", label, text)
					callback(text, nil)
					CloseModalNotification(app, pages, modalName, mainView)
					done()
				}
			}).
			AddButton("Skip", func() {
				callback("", fmt.Errorf("input skipped input %s", label))
				logger.Info("User skipped input %s", label)
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			}).
			AddButton("Cancel All", func() {
				cancelAllCallback()
				logger.Info("User canceled all inputs")
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			})

		// Center align the buttons
		form.SetButtonsAlign(tview.AlignCenter)
		form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)

		// Calculate position for centering
		modalX := (width - modalWidth) / 2
		modalY := (height - 7) / 2 // 7 is approximate height of the form

		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, modalY, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, modalX, 1, false).
				AddItem(form, modalWidth, 1, true).
				AddItem(nil, modalX, 1, false),
				7, 1, true).
			AddItem(nil, modalY, 1, false)

		pages.AddPage(modalName, flex, true, true)

		app.SetFocus(input)
		SetFloatingBoxActive(true)
	})
}
