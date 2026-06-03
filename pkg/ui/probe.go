package ui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/kartverket/skipctl/pkg/test"
	"github.com/rivo/tview"
)

func NewProbeView(servers []discovery.APIServer, activeServer *discovery.APIServer) tview.Primitive {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	results := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	results.SetBorder(true).SetTitle("Probe Results")

	form := tview.NewForm()

	serverOptions := make([]string, len(servers))
	initialServerIndex := 0
	for i, s := range servers {
		serverOptions[i] = s.Name
		if activeServer != nil && s.Name == activeServer.Name {
			initialServerIndex = i
		}
	}

	var selectedServer *discovery.APIServer
	if len(servers) > 0 {
		selectedServer = &servers[initialServerIndex]
	}

	form.AddDropDown("API Server", serverOptions, initialServerIndex, func(option string, optionIndex int) {
		selectedServer = &servers[optionIndex]
	})

	form.AddInputField("Hostname", "", 40, nil, nil)
	form.AddInputField("Port", "", 10, nil, nil)

	form.AddButton("Probe", func() {
		hostname := form.GetFormItemByLabel("Hostname").(*tview.InputField).GetText()
		portStr := form.GetFormItemByLabel("Port").(*tview.InputField).GetText()

		if selectedServer == nil {
			fmt.Fprintf(results, "[red]Error: No API server selected[white]\n")
			return
		}
		if hostname == "" {
			fmt.Fprintf(results, "[red]Error: Hostname is required[white]\n")
			return
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Fprintf(results, "[red]Error: Invalid port: %v[white]\n", err)
			return
		}

		fmt.Fprintf(results, "Probing [yellow]%s:%d[white] via [blue]%s[white]...\n", hostname, port, selectedServer.Name)
		results.ScrollToEnd()

		go func() {
			t, err := test.NewTester(context.Background(), selectedServer.Addr, true) // assuming TLS=true for now
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(results, "[red]Error creating tester: %v[white]\n", err)
				})
				return
			}

			res, err := t.PortProbe(context.Background(), hostname, int32(port), constants.DefaultTestTimeout)
			app.QueueUpdateDraw(func() {
				if err != nil {
					fmt.Fprintf(results, "[red]Probe failed: %v[white]\n", err)
				} else {
					status := "[red]CLOSED"
					if res.GetOpen() {
						status = "[green]OPEN"
					}
					fmt.Fprintf(results, "Result: %s[white] (Address probed: %s)\n", status, res.GetAddrProbed())
				}
				results.ScrollToEnd()
			})
		}()
	})

	form.AddButton("Clear", func() {
		results.Clear()
	})

	form.SetBorder(true).SetTitle("Port Probe Tool").SetTitleAlign(tview.AlignLeft)

	// Focus management
	focusList := []tview.Primitive{form, results}
	currentFocus := 0

	switchFocus := func() {
		currentFocus = (currentFocus + 1) % len(focusList)
		app.SetFocus(focusList[currentFocus])
	}

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			switchFocus()
			return nil
		}
		return event
	})

	results.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyCtrlN {
			switchFocus()
			return nil
		}
		return event
	})

	// Add visual feedback for focus
	form.SetFocusFunc(func() {
		form.SetBorderColor(tcell.ColorYellow)
	})
	form.SetBlurFunc(func() {
		form.SetBorderColor(tcell.ColorWhite)
	})
	results.SetFocusFunc(func() {
		results.SetBorderColor(tcell.ColorYellow)
	})
	results.SetBlurFunc(func() {
		results.SetBorderColor(tcell.ColorWhite)
	})

	flex.AddItem(form, 0, 1, true)
	flex.AddItem(results, 0, 2, false)

	return flex
}
