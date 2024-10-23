package functions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fortifyde/netutil/internal/functions/parsers"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/pkg"
)

// analyze tshark capture and return structured results
func AnalyzeTsharkCapture(captureFile string, analysisDir string) ([]pkg.AnalysisResult, error) {
	logger.Info("Starting tshark capture analysis")

	analysisCmds := GetAnalysisCommands(captureFile)

	var results []pkg.AnalysisResult

	for _, cmd := range analysisCmds {
		logger.Info("Running tshark command for: %s", cmd.Name)

		tsharkCmd := exec.Command("tshark", cmd.Args...)

		// capture stdout and stderr separately
		stdoutPipe, err := tsharkCmd.StdoutPipe()
		if err != nil {
			logger.Error("Failed to get stdout pipe for %s: %v", cmd.Name, err)
			continue
		}

		stderrPipe, err := tsharkCmd.StderrPipe()
		if err != nil {
			logger.Error("Failed to get stderr pipe for %s: %v", cmd.Name, err)
			continue
		}

		// start tshark command
		if err := tsharkCmd.Start(); err != nil {
			logger.Error("Failed to start tshark for %s: %v", cmd.Name, err)
			continue
		}

		go func(name string, stderr io.ReadCloser) {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				logger.Info("tshark [%s] stderr: %s", name, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				logger.Error("Error reading stderr for %s: %v", name, err)
			}
		}(cmd.Name, stderrPipe)

		outputBytes, err := io.ReadAll(stdoutPipe)
		if err != nil {
			logger.Error("Failed to read stdout for %s: %v", cmd.Name, err)
			continue
		}

		if err := tsharkCmd.Wait(); err != nil {
			logger.Error("tshark command failed for %s: %v", cmd.Name, err)
			continue
		}

		output := string(outputBytes)
		parsedOutput, err := parsers.ParseAnalysisOutput(cmd.Name, output)
		if err != nil {
			logger.Error("Failed to parse output for %s: %v", cmd.Name, err)
			continue
		}

		results = append(results, pkg.AnalysisResult{
			Name:   cmd.Name,
			Output: parsedOutput,
		})

		saveOutputs(cmd, analysisDir, output, parsedOutput)
	}

	createHTMLIndex(analysisCmds, analysisDir)

	return results, nil
}

// save raw and parsed outputs to files
func saveOutputs(cmd AnalysisCommand, analysisDir, rawOutput string, parsedOutput interface{}) {
	timestamp := time.Now().Format("20060102_150405")
	escapedName := strings.ReplaceAll(cmd.Name, " ", "_")

	// save raw
	rawFile := filepath.Join(analysisDir, fmt.Sprintf("%s_%s.txt", escapedName, timestamp))
	if err := os.WriteFile(rawFile, []byte(rawOutput), 0644); err != nil {
		logger.Error("Failed to write raw output for %s: %v", cmd.Name, err)
	} else {
		logger.Info("Raw output saved to %s", rawFile)
	}

	// save JSON
	jsonData, err := json.MarshalIndent(parsedOutput, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal JSON for %s: %v", cmd.Name, err)
		return
	}

	jsonFile := filepath.Join(analysisDir, fmt.Sprintf("%s_%s.json", escapedName, timestamp))
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		logger.Error("Failed to write JSON output for %s: %v", cmd.Name, err)
	} else {
		logger.Info("JSON output saved to %s", jsonFile)
	}
}

// create HTML index linking all analysis results
func createHTMLIndex(analysisCmds []AnalysisCommand, analysisDir string) {
	timestamp := time.Now().Format("20060102_150405")
	indexFile := filepath.Join(analysisDir, fmt.Sprintf("index_%s.html", timestamp))
	f, err := os.Create(indexFile)
	if err != nil {
		logger.Error("Failed to create index file: %v", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "<html><body><h1>Analysis Results</h1><ul>")
	for _, cmd := range analysisCmds {
		escapedName := strings.ReplaceAll(cmd.Name, " ", "_")
		fmt.Fprintf(f, "<li><a href='%s_%s.txt'>%s Raw Output</a></li>", escapedName, timestamp, cmd.Name)
		fmt.Fprintf(f, "<li><a href='%s_%s.json'>%s JSON</a></li>", escapedName, timestamp, cmd.Name)
	}
	fmt.Fprintf(f, "</ul></body></html>")
	logger.Info("Analysis index created at %s", indexFile)
}
