package uiutil

import (
	"fmt"

	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowAnalysisResults displays the analysis results in a modal with structured columns and merged headers.
func ShowAnalysisResults(app *tview.Application, pages *tview.Pages, modalName string, results []pkg.AnalysisResult, mainView tview.Primitive) {
	table := tview.NewTable().
		SetBorders(true).
		SetBordersColor(pkg.NordAccent).
		SetSelectable(false, false)

	// Define analysis categories and their sub-fields dynamically
	type Category struct {
		Name      string
		SubFields []string
		Data      interface{}
	}

	var categories []Category

	// Mapping between analysis names and their subfields
	categoryMapping := map[string][]string{
		"Network Topology":      {"IP", "Packet Count"},
		"VLAN IDs":              {"VLAN ID"},
		"Protocol Distribution": {"Protocol", "Frame Count"},
		"Unusual Protocols":     {"Protocol", "Packet Count"},
		"Unencrypted Protocols": {"Protocol", "Packet Count"},
		"Weak SSLTLS Versions":  {"Source IP", "Destination IP", "TLS Handshake Ciphersuite", "TLS Record Version"},
		// Add other categories and their subfields here...
	}

	// Populate categories based on results
	for _, result := range results {
		if subfields, exists := categoryMapping[result.Name]; exists {
			cat := Category{
				Name:      result.Name,
				SubFields: subfields,
				Data:      result.Output,
			}
			categories = append(categories, cat)
		}
	}

	if len(categories) == 0 {
		ShowMessage(app, pages, "noAnalysisResultsModal", "No analysis results to display.", mainView)
		return
	}

	// Calculate the total number of columns based on subfields
	totalColumns := 0
	for _, cat := range categories {
		totalColumns += len(cat.SubFields)
	}

	// Determine the maximum number of rows required
	maxRows := 0
	for _, cat := range categories {
		dataLength := 0
		switch data := cat.Data.(type) {
		case []pkg.NetworkTopology:
			dataLength = len(data)
		case []pkg.VLANID:
			dataLength = len(data)
		case []pkg.ProtocolDistribution:
			dataLength = len(data)
		case []pkg.UnusualProtocol:
			dataLength = len(data)
		case []pkg.UnencryptedProtocol:
			dataLength = len(data)
		case []pkg.WeakSSLTLS:
			dataLength = len(data)
		// Handle other data types similarly...
		default:
			dataLength = 0
		}
		if dataLength > maxRows {
			maxRows = dataLength
		}
	}

	// Populate header rows

	// Row 0: Category Names (simulated merged headers)
	col := 0
	for _, cat := range categories {
		table.SetCell(0, col, tview.NewTableCell(fmt.Sprintf("[::b]%s", cat.Name)).
			SetTextColor(pkg.NordFg).
			SetAlign(tview.AlignCenter).
			SetExpansion(len(cat.SubFields)).
			SetSelectable(false))
		// Fill the remaining cells under the merged header with empty cells
		for i := 1; i < len(cat.SubFields); i++ {
			table.SetCell(0, col+i, tview.NewTableCell("")).
				SetSelectable(false, false)
		}
		col += len(cat.SubFields)
	}

	// Row 1: Subfield Names
	col = 0
	for _, cat := range categories {
		for _, sub := range cat.SubFields {
			table.SetCell(1, col, tview.NewTableCell(fmt.Sprintf("[::b]%s", sub)).
				SetTextColor(pkg.NordFg).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
			col++
		}
	}

	// Populate data rows
	for row := 0; row < maxRows; row++ {
		col = 0
		for _, cat := range categories {
			switch data := cat.Data.(type) {
			case []pkg.NetworkTopology:
				if row < len(data) {
					table.SetCell(row+2, col, tview.NewTableCell(data[row].IP).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+1, tview.NewTableCell(fmt.Sprintf("%d", data[row].PacketCount)).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
				} else {
					// Display "No data available" spanning subfields
					table.SetCell(row+2, col, tview.NewTableCell("").
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignCenter).
						SetExpansion(len(cat.SubFields)).
						SetSelectable(false))
					break // Only need to set once per row for this category

				}
			case []pkg.VLANID:
				if row < len(data) {
					table.SetCell(row+2, col, tview.NewTableCell(data[row].VLANID).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
				} else {
					// Display "No data available"
					table.SetCell(row+2, col, tview.NewTableCell("").
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignCenter))
				}
			case []pkg.ProtocolDistribution:
				if row < len(data) {
					table.SetCell(row+2, col, tview.NewTableCell(data[row].Protocol).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+1, tview.NewTableCell(fmt.Sprintf("%d", data[row].FrameCount)).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
				} else {
					table.SetCell(row+2, col, tview.NewTableCell("").
						SetSelectable(false))
					table.SetCell(row+2, col+1, tview.NewTableCell("").
						SetSelectable(false))
				}
			case []pkg.UnusualProtocol:
				if row < len(data) {
					table.SetCell(row+2, col, tview.NewTableCell(data[row].Protocol).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+1, tview.NewTableCell(fmt.Sprintf("%d", data[row].PacketCount)).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
				} else {
					table.SetCell(row+2, col, tview.NewTableCell("").
						SetSelectable(false))
					table.SetCell(row+2, col+1, tview.NewTableCell("").
						SetSelectable(false))
				}
			case []pkg.UnencryptedProtocol:
				if row < len(data) {
					table.SetCell(row+2, col, tview.NewTableCell(data[row].Protocol).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+1, tview.NewTableCell(fmt.Sprintf("%d", data[row].PacketCount)).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
				} else {
					table.SetCell(row+2, col, tview.NewTableCell("").
						SetSelectable(false))
					table.SetCell(row+2, col+1, tview.NewTableCell("").
						SetSelectable(false))
				}
			case []pkg.WeakSSLTLS:
				if row < len(data) {
					table.SetCell(row+2, col, tview.NewTableCell(data[row].SourceIP).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+1, tview.NewTableCell(data[row].DestinationIP).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+2, tview.NewTableCell(data[row].TLSHandshakeCS).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
					table.SetCell(row+2, col+3, tview.NewTableCell(data[row].TLSRecordVersion).
						SetTextColor(pkg.NordFg).
						SetAlign(tview.AlignLeft))
				} else {
					table.SetCell(row+2, col, tview.NewTableCell("").
						SetSelectable(false))
					table.SetCell(row+2, col+1, tview.NewTableCell("").
						SetSelectable(false))
				}
			// Handle other data types similarly...
			default:
				// For unhandled categories, fill with "No data available"
				table.SetCell(row+2, col, tview.NewTableCell("No data available").
					SetSelectable(false))
			}
			// Move to next set of columns
			col += len(cat.SubFields)
		}
	}

	// Instruction TextView
	instruction := tview.NewTextView().
		SetText("Press 'q' to close").
		SetTextAlign(tview.AlignCenter).
		SetTextColor(pkg.NordHighlight)

	// Capture 'q' key to close the modal
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			CloseModalNotification(app, pages, "analysisResultsModal", mainView)
			return nil
		}
		return event
	})

	// Frame for the table
	frame := tview.NewFrame(table).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("Analysis Results", true, tview.AlignCenter, pkg.NordHighlight)

	// Combine frame and instruction into a Flex layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(frame, 0, 1, true).
		AddItem(instruction, 1, 1, false)

	pages.AddPage(modalName, flex, true, true)
	app.SetFocus(table)
	SetFloatingBoxActive(true)
}
