package ui

import (
	"github.com/rivo/tview"
)

func NewMenuView() tview.Primitive {
	list := tview.NewList().
		AddItem("Port Probe", "Check if a TCP port is open from a SKIP cluster", 'p', func() {
			pages.SwitchToPage("probe")
		}).
		AddItem("Ping Host", "Check if a host is responsive to ping from a SKIP cluster", 'i', func() {
			pages.SwitchToPage("ping")
		}).
		AddItem("Validate Manifests", "Validate Kubernetes manifests against schemas", 'v', func() {
			pages.SwitchToPage("validate")
		}).
		AddItem("Quit", "Exit the application", 'q', func() {
			app.Stop()
		})

	list.SetBorder(true).SetTitle("Main Menu")
	return list
}
