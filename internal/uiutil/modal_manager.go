package uiutil

import (
	"log"
	"sync"

	"github.com/rivo/tview"
)

// defines the signature for modal functions
// accepts a 'done' callback to signal when the modal has been closed
type ModalFunc func(done func())

// manages the display of modals in a queued manner
type ModalManager struct {
	app        *tview.Application
	pages      *tview.Pages
	mainView   tview.Primitive
	queue      []ModalFunc
	queueMutex sync.Mutex
	isActive   bool
}

// initializes and returns a new ModalManager
func NewModalManager(app *tview.Application, pages *tview.Pages, mainView tview.Primitive) *ModalManager {
	return &ModalManager{
		app:      app,
		pages:    pages,
		mainView: mainView,
		queue:    []ModalFunc{},
		isActive: false,
	}
}

// adds a new modal function to the queue
// logs a fatal error if the ModalManager is not initialized
func (m *ModalManager) Enqueue(modal ModalFunc) {
	if m == nil {
		log.Fatal("ModalManager is not initialized. Please call InitializeModalManager before enqueuing modals.")
	}

	m.queueMutex.Lock()
	defer m.queueMutex.Unlock()

	m.queue = append(m.queue, modal)
	log.Printf("Enqueued a new modal. Queue length: %d", len(m.queue))
	if !m.isActive {
		m.isActive = true
		go m.displayNext()
	}
}

// displays the next modal in the queue
func (m *ModalManager) displayNext() {
	for {
		m.queueMutex.Lock()
		if len(m.queue) == 0 {
			m.isActive = false
			m.queueMutex.Unlock()
			log.Println("No more modals in the queue. ModalManager is now inactive.")
			return
		}
		nextModal := m.queue[0]
		m.queue = m.queue[1:]
		m.queueMutex.Unlock()

		log.Println("Displaying the next modal.")

		done := make(chan struct{})
		m.app.QueueUpdateDraw(func() {
			nextModal(func() {
				log.Println("Modal has been closed.")
				close(done)
			})
		})

		<-done
	}
}
