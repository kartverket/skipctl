package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/rivo/tview"
)

var (
	app    *tview.Application
	pages  *tview.Pages
	header *tview.TextView
	footer *tview.TextView
)

func Run(discoveryHost, apiServerName string) error {
	logging.Mute()

	servers, err := discovery.DiscoverAPIServers(discoveryHost)
	if err != nil {
		return fmt.Errorf("failed to discover API servers: %w", err)
	}

	var activeServer *discovery.APIServer
	if apiServerName != "" {
		for _, s := range servers {
			if s.Name == apiServerName {
				activeServer = &s
				break
			}
		}
	}

	app = tview.NewApplication()
	pages = tview.NewPages()

	header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]SKIPCTL[white]")

	footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[darkgray]Esc: Menu | Ctrl-C: Quit[white]")

	// Initialize Views
	menuView := NewMenuView()
	probeView := NewProbeView(servers, activeServer)
	pingView := NewPingView(servers, activeServer)
	validateView := NewValidateView()

	pages.AddPage("menu", menuView, true, true)
	pages.AddPage("probe", probeView, true, false)
	pages.AddPage("ping", pingView, true, false)
	pages.AddPage("validate", validateView, true, false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 1, false).
		AddItem(pages, 0, 1, true).
		AddItem(footer, 1, 1, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		return err
	}

	return nil
}
