package uiutil

import (
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

// OutputModal represents the structure for the output modal
type OutputModal struct {
	textView *tview.TextView
	mutex    sync.Mutex
	app      *tview.Application
	pages    *tview.Pages
	name     string
	doneFunc func()
}

// ShowOutputModal enqueues and displays the output modal with adjusted size.
func ShowOutputModal(app *tview.Application, pages *tview.Pages, modalName, title string, onCancel func()) *OutputModal {
	modalName = "output_" + modalName
	outputModal := &OutputModal{
		textView: tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			SetChangedFunc(func() {
				app.Draw()
			}),
		app:   app,
		pages: pages,
		name:  modalName,
	}

	// Configure the TextView
	outputModal.textView.
		SetWrap(false). // Disable wrapping to maintain terminal-like output
		ScrollToEnd().SetScrollable(true).
		SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	// Get terminal dimensions using x/term
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fallback to reasonable defaults if we can't get terminal size
		width, height = 100, 30
	}

	// Calculate modal size (70% width, 80% height)
	modalWidth := int(float64(width) * 0.7)
	modalHeight := int(float64(height) * 0.8)

	// Calculate position for centering
	modalX := (width - modalWidth) / 2
	modalY := (height - modalHeight) / 2

	// Define the layout with fixed size
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, modalY, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, modalX, 1, false).
			AddItem(outputModal.textView, modalWidth, 1, true).
			AddItem(nil, modalX, 1, false), modalHeight, 1, true).
		AddItem(nil, modalY, 1, false)

	// Add the modal to pages
	pages.AddPage(modalName, flex, true, true)

	// Set focus to the TextView
	app.SetFocus(outputModal.textView)

	// Handle user input for cancellation (Ctrl+C)
	outputModal.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			if onCancel != nil {
				onCancel()
			}
			return nil
		}
		return event
	})

	// Mark the floating box as active
	SetFloatingBoxActive(true)

	// Enqueue the modal using ModalManager
	modalManager.Enqueue(func(done func()) {
		outputModal.doneFunc = done
	})

	return outputModal
}

// AppendText safely appends text to the OutputModal's TextView.
func (o *OutputModal) AppendText(text string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.app.QueueUpdateDraw(func() {
		o.textView.Write([]byte(text))
	})
}

// CloseOutputModal gracefully closes the output modal with an "OK" button.
func (o *OutputModal) CloseOutputModal(completionMessage string) {
	o.app.QueueUpdateDraw(func() {
		// Create a new modal with the completion message and "OK" button
		modal := tview.NewModal().
			SetText(completionMessage).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				// Remove the output modal
				o.pages.RemovePage(o.name)
				o.pages.RemovePage(o.name + "_completion")
				SetFloatingBoxActive(false)
				// Enqueue the closure
				if o.doneFunc != nil {
					o.doneFunc()
				}
			})

		// Define the layout for the completion modal
		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modal, 5, 1, true).
			AddItem(nil, 0, 1, false)

		o.pages.AddPage(o.name+"_completion", flex, true, true)
		o.app.SetFocus(modal)
	})
}
