package scripts

import (
	"log"
	"os/exec"

	"github.com/rivo/tview"
)

func RunBashScript(scriptPath string, toolbox *tview.TextView) {
	toolbox.SetText("Running script: " + scriptPath + "\n") // placeholder
	cmd := exec.Command("bash", scriptPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				_, err := toolbox.Write(buf[:n])
				if err != nil {
					log.Printf("Failed to write to toolbox: %v", err)
				}
			}
			if err != nil {
				break
			}
		}
	}()
}
