package uiutil

import (
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"

	"github.com/fortifyde/netutil/internal/logger"
)

// OutputModal represents the structure for the output modal
type OutputModal struct {
	textView         *tview.TextView
	mutex            sync.Mutex
	app              *tview.Application
	pages            *tview.Pages
	name             string
	doneFunc         func()
	isScanning       bool                                  // Flag to track if scanning is in progress
	once             sync.Once                             // Ensures doneFunc is called only once
	prevFocus        tview.Primitive                       // Add this field
	prevInputCapture func(*tcell.EventKey) *tcell.EventKey // Add this field
}

// SetScanning sets the scanning state of the OutputModal
func (o *OutputModal) SetScanning(scanning bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.isScanning = scanning
}

// IsScanning returns the current scanning state
func (o *OutputModal) IsScanning() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.isScanning
}

// ShowOutputModal enqueues and displays the output modal with adjusted size.
// It uses ModalManager to manage the modal display.
// Returns a pointer to the OutputModal for appending text.
func ShowOutputModal(app *tview.Application, pages *tview.Pages, modalName, title string, onCancel func(), mainView tview.Primitive) *OutputModal {
	var outputModal *OutputModal
	var initWG sync.WaitGroup
	initWG.Add(1) // Ensure that the caller waits until OutputModal is initialized

	// Store the previous input capture
	previousInputCapture := app.GetInputCapture()

	modalFunc := func(done func()) {
		// Mark the floating box as active
		SetFloatingBoxActive(true)
		// Initialize OutputModal
		outputModal = &OutputModal{
			textView: tview.NewTextView().
				SetDynamicColors(true).
				SetRegions(true).
				SetScrollable(true).
				SetChangedFunc(func() {
					app.Draw()
				}),
			app:              app,
			pages:            pages,
			name:             modalName,
			isScanning:       true, // Initialize as scanning in progress
			doneFunc:         done,
			prevFocus:        mainView,
			prevInputCapture: previousInputCapture, // Store the previous input capture
		}

		// Configure the TextView
		outputModal.textView.
			SetWrap(false). // Disable wrapping to maintain terminal-like output
			ScrollToEnd().SetBorder(true).
			SetTitle(title).
			SetTitleAlign(tview.AlignLeft)

		// Set up input capture at the application level
		outputModal.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

			switch event.Key() {
			case tcell.KeyCtrlC:
				if onCancel != nil && outputModal.IsScanning() {
					logger.Info("OutputModal: Ctrl+C pressed")
					onCancel()
					return nil
				}
			case tcell.KeyEnter:
				if !outputModal.IsScanning() {
					outputModal.CloseModal()
					return nil
				}
			}

			if event.Rune() == 'q' || event.Rune() == 'Q' {
				if onCancel != nil && outputModal.IsScanning() {
					logger.Info("OutputModal: 'q' pressed")
					onCancel()
					return nil
				}
			}

			return event
		})
		// Create instruction text
		instructions := tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText("[yellow]Press Ctrl+C or 'q' to cancel operation. Enter to close after completion of scans.[-]").
			SetDynamicColors(true)

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

		// Define the layout with fixed size and instructions
		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, modalY, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, modalX, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(outputModal.textView, modalHeight-1, 1, true).
					AddItem(instructions, 1, 0, false),
					modalWidth, 1, true).
				AddItem(nil, modalX, 1, false),
				modalHeight, 1, true).
			AddItem(nil, modalY, 1, false)

		// Add the modal to pages
		pages.AddPage(modalName, flex, true, true)
		outputModal.SetScanning(true)

		// Set focus to the TextView
		outputModal.app.SetFocus(outputModal.textView)

		// Signal that OutputModal has been initialized
		initWG.Done()
	}

	// Enqueue the ModalFunc using ModalManager
	modalManager := GetModalManager()
	if modalManager != nil {
		modalManager.Enqueue(modalFunc)
	} else {
		// Handle error: ModalManager not initialized
		panic("ModalManager is not initialized. Please initialize it before showing modals.")
	}

	// Wait until OutputModal is initialized
	initWG.Wait()

	return outputModal
}

// AppendText safely appends text to the OutputModal's TextView.
func (o *OutputModal) AppendText(text string) {
	o.app.QueueUpdateDraw(func() {
		o.textView.Write([]byte(text))
	})
}

// CloseModal gracefully closes the output modal and restores previous state
func (o *OutputModal) CloseModal() {
	o.once.Do(func() {
		go o.app.QueueUpdateDraw(func() {
			o.SetScanning(false)
			o.pages.RemovePage(o.name)
			SetFloatingBoxActive(false)

			// Restore the previous input capture first
			if o.prevInputCapture != nil {
				o.app.SetInputCapture(o.prevInputCapture)
			}

			// Then restore focus
			if toolbox, ok := o.prevFocus.(*tview.List); ok {
				o.app.SetFocus(toolbox)
			} else if o.prevFocus != nil {
				o.app.SetFocus(o.prevFocus)
			}

			if o.doneFunc != nil {
				o.doneFunc()
			}
		})
	})
}

// CloseOutputModal gracefully closes the output modal with a completion message.
func (o *OutputModal) CloseOutputModal(completionMessage string) {
	o.AppendText(completionMessage + "\n")
	o.CloseModal()
}
