package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
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
	pathInput := tview.NewInputField().
		SetLabel("Path").
		SetText(".").
		SetFieldWidth(40)
	form.AddFormItem(pathInput)

	// Tree View
	tree := tview.NewTreeView().
		SetRoot(tview.NewTreeNode(".").
			SetColor(tcell.ColorGreen).
			SetReference("."))
	tree.SetBorder(true).SetTitle("Directory Explorer")

	// Helper to add nodes
	addNodes := func(target *tview.TreeNode, path string, limit int) {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return
		}
		count := 0
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") && file.Name() != "." {
				continue
			}
			if file.Name() == "node_modules" || file.Name() == ".git" {
				continue
			}

			isManifest := false
			for _, suffix := range constants.ManifestSuffixes {
				if strings.HasSuffix(file.Name(), suffix) {
					isManifest = true
					break
				}
			}

			if !file.IsDir() && !isManifest {
				continue
			}

			if limit > 0 && count >= limit {
				target.AddChild(tview.NewTreeNode("...").SetSelectable(false))
				break
			}
			node := tview.NewTreeNode(file.Name()).
				SetReference(filepath.Join(path, file.Name()))
			if file.IsDir() {
				node.SetColor(tcell.ColorGreen)
			}
			target.AddChild(node)
			count++
		}
	}

	updateTree := func(path string) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return
		}

		info, err := os.Stat(absPath)
		if err != nil || !info.IsDir() {
			return
		}

		root := tview.NewTreeNode(absPath).
			SetColor(tcell.ColorGreen).
			SetReference(absPath)
		tree.SetRoot(root)
		tree.SetCurrentNode(root)

		if absPath == "/" {
			// Special handling for root to avoid scanning everything
			addNodes(root, "/", 20)
		} else {
			addNodes(root, absPath, 0)
		}
		root.SetExpanded(true)
	}

	// Update tree only when user presses Enter in the path input
	pathInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			updateTree(pathInput.GetText())
		}
	})

	// Initial load
	updateTree(".")

	// Real-time path update as we navigate the tree
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		path := reference.(string)
		// Update path input without triggering a tree reload
		pathInput.SetText(path)
	})

	runValidation := func(path string) {
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
	}

	toggleNode := func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		path := reference.(string)

		stat, err := os.Stat(path)
		if err != nil {
			return
		}

		if stat.IsDir() {
			if node == tree.GetRoot() {
				// Toggle expansion for root
				children := node.GetChildren()
				if len(children) == 0 {
					addNodes(node, path, 0)
				}
				node.SetExpanded(!node.IsExpanded())
			} else {
				// Enter directory (make it the new root)
				updateTree(path)
			}
		} else {
			runValidation(path)
		}
	}

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		toggleNode(node)
	})

	// Focus management
	focusList := []tview.Primitive{tree, form, results}
	currentFocus := 0

	updateFooter := func() {
		if footer != nil {
			footer.SetText("[darkgray]Tab: Next Pane | Esc: Menu | Backspace: Go Up | Enter: Enter Dir | Ctrl-N/P: Switch Panes[white]")
		}
	}

	switchFocus := func(next bool) {
		if next {
			currentFocus = (currentFocus + 1) % len(focusList)
		} else {
			currentFocus = (currentFocus - 1 + len(focusList)) % len(focusList)
		}
		app.SetFocus(focusList[currentFocus])
		updateFooter()
	}

	goUp := func() {
		root := tree.GetRoot()
		if root != nil {
			reference := root.GetReference()
			if reference != nil {
				path := reference.(string)
				parent := filepath.Dir(path)
				if parent != path {
					updateTree(parent)
				}
			}
		}
	}

	// Shared input capture for switching panes
	paneCapture := func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN:
			switchFocus(true)
			return nil
		case tcell.KeyCtrlP:
			switchFocus(false)
			return nil
		}
		return event
	}

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			switchFocus(true)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			switchFocus(false)
			return nil
		}

		// Fallback to pane capture for Ctrl-N/P
		if e := paneCapture(event); e == nil {
			return nil
		}

		node := tree.GetCurrentNode()
		if node == nil {
			return event
		}

		switch event.Key() {
		case tcell.KeyRight:
			reference := node.GetReference()
			if reference == nil {
				return event
			}
			path := reference.(string)
			stat, err := os.Stat(path)
			if err == nil && stat.IsDir() {
				children := node.GetChildren()
				if len(children) == 0 {
					addNodes(node, path, 0)
				}
				node.SetExpanded(true)
				return nil
			}
		case tcell.KeyLeft:
			if node.IsExpanded() {
				node.SetExpanded(false)
				return nil
			}
			goUp()
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			goUp()
			return nil
		}
		return event
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return paneCapture(event)
	})

	results.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			switchFocus(true)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			switchFocus(false)
			return nil
		}
		return paneCapture(event)
	})

	form.AddButton("Validate", func() {
		runValidation(pathInput.GetText())
	})

	form.AddButton("Clear Results", func() {
		results.Clear()
	})

	form.SetBorder(true).SetTitle("Validate Manifests Tool").SetTitleAlign(tview.AlignLeft)

	// Add visual feedback for focus
	tree.SetFocusFunc(func() {
		tree.SetBorderColor(tcell.ColorYellow)
		updateFooter()
	})
	tree.SetBlurFunc(func() {
		tree.SetBorderColor(tcell.ColorWhite)
	})
	form.SetFocusFunc(func() {
		form.SetBorderColor(tcell.ColorYellow)
		updateFooter()
	})
	form.SetBlurFunc(func() {
		form.SetBorderColor(tcell.ColorWhite)
	})
	results.SetFocusFunc(func() {
		results.SetBorderColor(tcell.ColorYellow)
		updateFooter()
	})
	results.SetBlurFunc(func() {
		results.SetBorderColor(tcell.ColorWhite)
	})

	// Side-by-side layout
	contentFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tree, 0, 1, true).
		AddItem(results, 0, 2, false)

	flex.AddItem(form, 7, 1, false)
	flex.AddItem(contentFlex, 0, 1, true)

	return flex
}

func getManifestFiles(path string) ([]*manifest.Document, error) {
	filenames, err := utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
	if err != nil {
		return nil, err
	}
	return manifest.FromFiles(filenames)
}
