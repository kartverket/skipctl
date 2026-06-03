package ui

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/rivo/tview"
)

func NewManifestsView() tview.Primitive {
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Sidebar
	sidebar := tview.NewFlex().SetDirection(tview.FlexRow)

	// Path and Results combined window
	pathAndResults := tview.NewFlex().SetDirection(tview.FlexRow)
	pathAndResults.SetBorder(true).SetTitle("Manifests")

	results := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	pathInput := tview.NewInputField().
		SetLabel("Path: ").
		SetText(".").
		SetFieldWidth(0)

	pathAndResults.AddItem(pathInput, 1, 1, true)

	// Tree View
	tree := tview.NewTreeView().
		SetRoot(tview.NewTreeNode(".").
			SetColor(tcell.ColorGreen).
			SetReference("."))
	tree.SetBorder(true).SetTitle("Directory Explorer")

	sidebar.AddItem(pathAndResults, 10, 1, true)
	sidebar.AddItem(tree, 0, 2, false)

	// Render View (The largest window)
	renderView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	renderView.SetBorder(true).SetTitle("Manifest Rendering")

	mainFlex.AddItem(sidebar, 0, 1, true)
	mainFlex.AddItem(renderView, 0, 3, false)

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

	// Real-time path update as we navigate the tree
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		path := reference.(string)
		pathInput.SetText(path)
	})

	runValidation := func(path string) {
		results.Clear()
		fmt.Fprintf(results, "Validating [yellow]%s[white]...\n", path)
		results.ScrollToEnd()

		go func() {
			manifestFiles, err := getManifestFiles(path)
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(results, "[red]Error: %v[white]\n", err)
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
					fmt.Fprintf(results, "[red]Error: %v[white]\n", err)
				})
				return
			}
			defer os.RemoveAll(tempDir)

			processor := manifest.NewDocumentProcessor()
			buf := &bytes.Buffer{}
			rawLogger := logging.NewRawLoggerTo(buf)
			validator := manifest.NewValidator(tempDir, false).WithLogger(rawLogger)

			err = processor.ProcessDocuments(manifestFiles, validator.ValidateManifest)

			result := validator.GetResults()
			app.QueueUpdateDraw(func() {
				results.Clear()
				if err != nil {
					fmt.Fprintf(results, "[red]Errors:[white] %v\n", err)
				}
				ansiWriter := tview.ANSIWriter(results)
				ansiWriter.Write(buf.Bytes())

				fmt.Fprintf(results, "\nTotal: [blue]%d[white] Valid: [green]%d[white] Invalid: [red]%d[white]",
					result.GetTotalResources(), result.ValidCount, result.InvalidCount)
				results.ScrollToEnd()
			})
		}()
	}

	runRendering := func(path string) {
		renderView.Clear()
		fmt.Fprintf(renderView, "Rendering [yellow]%s[white]...\n", path)

		go func() {
			docs, err := getManifestFiles(path)
			if err != nil {
				app.QueueUpdateDraw(func() {
					renderView.Clear()
					fmt.Fprintf(renderView, "[red]Error reading path: %v[white]\n", err)
				})
				return
			}

			if len(docs) == 0 {
				app.QueueUpdateDraw(func() {
					renderView.Clear()
					fmt.Fprintf(renderView, "[yellow]No manifests found in %s.[white]\n", path)
				})
				return
			}

			outputBuf := &bytes.Buffer{}

			for i, doc := range docs {
				if i > 0 {
					fmt.Fprintln(outputBuf)
					fmt.Fprintln(outputBuf, "========================================")
					fmt.Fprintln(outputBuf)
				}

				fmt.Fprintf(outputBuf, "[blue]Manifest:[white] %s\n", doc.Name)

				renderBuf := &bytes.Buffer{}
				rawLogger := logging.NewRawLoggerTo(renderBuf)
				renderer := manifest.NewRenderer(rawLogger)

				err = renderer.Render(doc)
				if err != nil {
					fmt.Fprintf(outputBuf, "[red]Error rendering: %v[white]\n", err)
					continue
				}

				renderedContent := renderBuf.String()

				// Validate rendered YAML content.
				tempDir, err := utils.CreateTempDirectory("schemas-render")
				if err != nil {
					fmt.Fprintf(outputBuf, "[red]Error creating temp dir: %v[white]\n", err)
					continue
				}

				validator := manifest.NewValidator(tempDir, false)
				valBuf := &bytes.Buffer{}
				validator.WithLogger(logging.NewRawLoggerTo(valBuf))

				renderedDoc := &manifest.Document{
					Name:      doc.Name,
					Content:   renderedContent,
					Extension: ".yaml", // Rendered output is always YAML
				}
				_ = validator.ValidateManifest(renderedDoc)
				result := validator.GetResults()
				_ = os.RemoveAll(tempDir)

				statusColor := "green"
				if result.HasValidationFailed() {
					statusColor = "red"
				}
				fmt.Fprintf(outputBuf, "[%s]Validation: %d Valid, %d Invalid, %d Error[white]\n",
					statusColor, result.ValidCount, result.InvalidCount, result.ErrorCount)

				if valBuf.Len() > 0 {
					outputBuf.Write(valBuf.Bytes())
					if !strings.HasSuffix(valBuf.String(), "\n") {
						fmt.Fprintln(outputBuf)
					}
				}

				fmt.Fprintln(outputBuf, "---")
				fmt.Fprint(outputBuf, renderedContent)
				if !strings.HasSuffix(renderedContent, "\n") {
					fmt.Fprintln(outputBuf)
				}
			}

			app.QueueUpdateDraw(func() {
				renderView.Clear()
				ansiWriter := tview.ANSIWriter(renderView)
				ansiWriter.Write(outputBuf.Bytes())
				renderView.ScrollToBeginning()
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
				children := node.GetChildren()
				if len(children) == 0 {
					addNodes(node, path, 0)
				}
				node.SetExpanded(!node.IsExpanded())
			} else {
				updateTree(path)
			}
		} else {
			runValidation(path)
			runRendering(path)
		}
	}

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		toggleNode(node)
	})

	pathInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			path := pathInput.GetText()
			updateTree(path)
			runValidation(path)

			stat, err := os.Stat(path)
			if err == nil && !stat.IsDir() {
				runRendering(path)
			}
		}
	})

	// Initial load
	updateTree(".")

	// Focus management
	focusList := []tview.Primitive{tree, pathInput, results, renderView}
	currentFocus := 0

	updateFooter := func() {
		if footer != nil {
			footer.SetText("[darkgray]Tab: Next Pane | Esc: Menu | Backspace: Go Up | Enter: Select/Dir | Ctrl-R: Validate+Render Path | Ctrl-N/P: Switch Panes[white]")
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

	paneCapture := func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlR:
			path := pathInput.GetText()
			if path != "" {
				runValidation(path)
				runRendering(path)
			}
			return nil
		case tcell.KeyCtrlN:
			switchFocus(true)
			return nil
		case tcell.KeyCtrlP:
			switchFocus(false)
			return nil
		case tcell.KeyTab:
			switchFocus(true)
			return nil
		case tcell.KeyBacktab:
			switchFocus(false)
			return nil
		}
		return event
	}

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

	pathInput.SetInputCapture(paneCapture)
	results.SetInputCapture(paneCapture)
	renderView.SetInputCapture(paneCapture)

	// Visual feedback for focus
	tree.SetFocusFunc(func() { tree.SetBorderColor(tcell.ColorYellow); updateFooter() })
	tree.SetBlurFunc(func() { tree.SetBorderColor(tcell.ColorWhite) })

	pathInput.SetFocusFunc(func() {
		pathAndResults.SetBorderColor(tcell.ColorYellow)
		updateFooter()
	})
	pathInput.SetBlurFunc(func() {
		pathAndResults.SetBorderColor(tcell.ColorWhite)
	})

	results.SetFocusFunc(func() {
		pathAndResults.SetBorderColor(tcell.ColorYellow)
		updateFooter()
	})
	results.SetBlurFunc(func() {
		pathAndResults.SetBorderColor(tcell.ColorWhite)
	})

	renderView.SetFocusFunc(func() { renderView.SetBorderColor(tcell.ColorYellow); updateFooter() })
	renderView.SetBlurFunc(func() { renderView.SetBorderColor(tcell.ColorWhite) })

	return mainFlex
}

func getManifestFiles(path string) ([]*manifest.Document, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return manifest.FromFiles([]string{path})
	}

	filenames, err := utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
	if err != nil {
		return nil, err
	}
	return manifest.FromFiles(filenames)
}
