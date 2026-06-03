package ui

import (
	"fmt"
	"os"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/rivo/tview"
)

func NewValidateView() tview.Primitive {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	results := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	results.SetBorder(true).SetTitle("Validation Results")

	form := tview.NewForm()
	form.AddInputField("Path", ".", 40, nil, nil)

	form.AddButton("Validate", func() {
		path := form.GetFormItemByLabel("Path").(*tview.InputField).GetText()

		fmt.Fprintf(results, "Validating manifests in [yellow]%s[white]...\n", path)
		results.ScrollToEnd()

		go func() {
			manifestFiles, err := getManifestFiles(path)
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(results, "[red]Error finding manifests: %v[white]\n", err)
				})
				return
			}

			if len(manifestFiles) == 0 {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(results, "[yellow]No manifests found.[white]\n")
				})
				return
			}

			tempDir, err := utils.CreateTempDirectory("schemas")
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(results, "[red]Error creating temp dir: %v[white]\n", err)
				})
				return
			}
			defer os.RemoveAll(tempDir)

			processor := manifest.NewDocumentProcessor()
			validator := manifest.NewValidator(tempDir, false)

			err = processor.ProcessDocuments(manifestFiles, validator.ValidateManifest)

			result := validator.GetResults()
			app.QueueUpdateDraw(func() {
				if err != nil {
					fmt.Fprintf(results, "[red]Errors occurred during validation:[white]\n%v\n\n", err)
				}
				fmt.Fprintf(results, "Validation completed:\n")
				fmt.Fprintf(results, "  Total resources: [blue]%d[white]\n", result.GetTotalResources())
				fmt.Fprintf(results, "  Valid:           [green]%d[white]\n", result.ValidCount)
				fmt.Fprintf(results, "  Invalid:         [red]%d[white]\n", result.InvalidCount)
				fmt.Fprintf(results, "  Errors:          [red]%d[white]\n", result.ErrorCount)
				fmt.Fprintf(results, "  Skipped:         [yellow]%d[white]\n", result.SkippedCount)
				results.ScrollToEnd()
			})
		}()
	})

	form.AddButton("Clear", func() {
		results.Clear()
	})

	form.SetBorder(true).SetTitle("Validate Manifests Tool").SetTitleAlign(tview.AlignLeft)

	flex.AddItem(form, 0, 1, true)
	flex.AddItem(results, 0, 2, false)

	return flex
}

func getManifestFiles(path string) ([]*manifest.Document, error) {
	filenames, err := utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
	if err != nil {
		return nil, err
	}
	return manifest.FromFiles(filenames)
}
