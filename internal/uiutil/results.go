package uiutil

import (
	"fmt"

	"github.com/fortifyde/netutil/internal/pkg"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// displays the analysis results in a modal with structured columns and merged headers
func ShowAnalysisResults(app *tview.Application, pages *tview.Pages, modalName string, results []pkg.AnalysisResult, mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
		table := tview.NewTable().
			SetBorders(true).
			SetBordersColor(pkg.NordAccent).
			SetSelectable(false, false)

		type Category struct {
			Name      string
			SubFields []string
			Data      interface{}
		}

		var categories []Category

		categoryMapping := map[string][]string{
			"Network Topology":      {"IP", "Packet Count"},
			"VLAN IDs":              {"VLAN ID"},
			"Protocol Distribution": {"Protocol", "Frame Count"},
			"Unusual Protocols":     {"Protocol", "Packet Count"},
			"Unencrypted Protocols": {"Protocol", "Packet Count"},
			"Weak SSLTLS Versions":  {"Source IP", "Destination IP", "TLS Handshake Ciphersuite", "TLS Record Version"},
		}

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
			done()
			return
		}

		totalColumns := 0
		for _, cat := range categories {
			totalColumns += len(cat.SubFields)
		}
		table.SetFixed(1, 1)

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
			default:
				dataLength = 0
			}
			if dataLength > maxRows {
				maxRows = dataLength
			}
		}

		col := 0
		for _, cat := range categories {
			table.SetCell(0, col, tview.NewTableCell(fmt.Sprintf("[::b]%s", cat.Name)).
				SetTextColor(pkg.NordFg).
				SetAlign(tview.AlignCenter).
				SetExpansion(len(cat.SubFields)).
				SetSelectable(false))
			for i := 1; i < len(cat.SubFields); i++ {
				table.SetCell(0, col+i, tview.NewTableCell("")).
					SetSelectable(false, false)
			}
			col += len(cat.SubFields)
		}

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
						table.SetCell(row+2, col, tview.NewTableCell("").
							SetTextColor(pkg.NordFg).
							SetAlign(tview.AlignCenter).
							SetExpansion(len(cat.SubFields)).
							SetSelectable(false))
						break
					}
				case []pkg.VLANID:
					if row < len(data) {
						table.SetCell(row+2, col, tview.NewTableCell(data[row].VLANID).
							SetTextColor(pkg.NordFg).
							SetAlign(tview.AlignLeft))
					} else {
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
						table.SetCell(row+2, col+2, tview.NewTableCell("").
							SetSelectable(false))
						table.SetCell(row+2, col+3, tview.NewTableCell("").
							SetSelectable(false))
					}
				default:
					table.SetCell(row+2, col, tview.NewTableCell("No data available").
						SetSelectable(false))
				}
				col += len(cat.SubFields)
			}
		}

		instruction := tview.NewTextView().
			SetText("Press 'q' to close").
			SetTextAlign(tview.AlignCenter).
			SetTextColor(pkg.NordHighlight)

		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'q', 'Q':
				CloseModalNotification(app, pages, modalName, mainView)
				done()
				return nil
			}
			return event
		})

		frame := tview.NewFrame(table).
			SetBorders(0, 0, 0, 0, 0, 0).
			AddText("Analysis Results", true, tview.AlignCenter, pkg.NordHighlight)

		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(frame, 0, 1, true).
			AddItem(instruction, 1, 1, false)

		pages.AddPage(modalName, flex, true, true)
		app.SetFocus(table)
		SetFloatingBoxActive(true)

		done()
	})
}

// displays the scan result in a modal
func ShowScanResult(app *tview.Application, pages *tview.Pages, modalName, result string, mainView tview.Primitive) {
	modalManager.Enqueue(func(done func()) {
		modal := tview.NewModal().
			SetText(result).
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				CloseModalNotification(app, pages, modalName, mainView)
				done()
			})

		pages.AddPage(modalName, modal, true, true)
		app.SetFocus(modal)
		SetFloatingBoxActive(true)
	})
}
