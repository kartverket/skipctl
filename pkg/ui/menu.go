package ui

import (
	"github.com/rivo/tview"
)

func NewMenuView() tview.Primitive {

	banner := tview.NewTextView().
		SetText(`
 ▄▄▄▄▄▄▄ ▄▄▄   ▄▄▄ ▄▄▄▄▄ ▄▄▄▄▄▄▄    ▄▄▄▄▄▄▄ ▄▄▄▄▄▄▄▄▄ ▄▄▄      
█████▀▀▀ ███ ▄███▀  ███  ███▀▀███▄ ███▀▀▀▀▀ ▀▀▀███▀▀▀ ███      
 ▀████▄  ███████    ███  ███▄▄███▀ ███         ███    ███      
   ▀████ ███▀███▄   ███  ███▀▀▀▀   ███         ███    ███      
███████▀ ███  ▀███ ▄███▄ ███       ▀███████    ███    ████████
`).
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetDynamicColors(false)
	list := tview.NewList().
		AddItem("Port Probe", "Check if a TCP port is open", 'p', func() {
			pages.SwitchToPage("probe")
		}).
		AddItem("Ping Host", "Check if a host is responsive", 'i', func() {
			pages.SwitchToPage("ping")
		}).
		AddItem("Validate Manifests", "Validate Kubernetes manifests", 'v', func() {
			pages.SwitchToPage("validate")
		}).
		AddItem("Quit", "Exit the application", 'q', func() {
			app.Stop()
		})

	list.SetBorder(true).SetTitle("Main Menu")

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(banner, 7, 1, false).
		AddItem(list, 0, 1, true)

	return layout
}
