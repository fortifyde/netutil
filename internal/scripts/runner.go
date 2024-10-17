package scripts

import (
	"log"
	"os/exec"

	"github.com/rivo/tview"
)

func RunBashScript(scriptPath string, outputBox *tview.TextView) {
	cmd := exec.Command("bash", scriptPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Capture and display output in the outputBox
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				outputBox.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()
}
