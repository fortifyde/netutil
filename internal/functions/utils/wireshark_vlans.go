package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
)

// VLANAnalysis represents the structure of the VLAN analysis JSON output
type VLANAnalysis []struct {
	VLANID string `json:"VLANID"`
}

// GetWiresharkVLANs reads the VLAN analysis results from the most recent Wireshark capture
func GetWiresharkVLANs() ([]string, error) {
	// Load config to get working directory
	cfg, err := configuration.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Construct path to analysis directory
	analysisDir := filepath.Join(cfg.WorkingDirectory, "Wireshark", "capture_analysis")

	// Find the most recent VLAN analysis JSON file
	matches, err := filepath.Glob(filepath.Join(analysisDir, "VLAN_IDs_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to search for VLAN analysis files: %v", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no VLAN analysis files found")
	}

	// Sort files by modification time to get the most recent
	sort.Slice(matches, func(i, j int) bool {
		iInfo, _ := os.Stat(matches[i])
		jInfo, _ := os.Stat(matches[j])
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	// Read the most recent file
	data, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read VLAN analysis file: %v", err)
	}

	// Parse the JSON
	var analysis VLANAnalysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse VLAN analysis: %v", err)
	}

	// Convert the VLAN objects to strings
	vlanIDs := make([]string, len(analysis))
	for i, vlan := range analysis {
		vlanIDs[i] = vlan.VLANID
	}

	logger.Info("Found %d VLANs in Wireshark analysis", len(vlanIDs))
	return vlanIDs, nil
}
