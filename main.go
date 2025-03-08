package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Initialize the application
	app := tview.NewApplication()

	// Create a tree view for displaying tests
	root := tview.NewTreeNode("Tests").
		SetColor(tcell.ColorYellow)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	// Create a text view for displaying test output
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)
	outputView.SetBorder(true).SetTitle("Test Output")

	// Discover pytest tests
	tests, err := discoverTests()
	if err != nil {
		fmt.Printf("Error discovering tests: %v\n", err)
		os.Exit(1)
	}

	// Add tests to the tree
	addTestsToTree(root, tests)

	// Create a status bar
	statusBar := tview.NewTextView().
		SetText("Ctrl+C: Exit | r: Run pytest").
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignCenter)

	// Create a horizontal flex layout for the tree and output view
	contentFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tree, 0, 1, true).
		AddItem(outputView, 0, 2, false) // Output view takes 2/3 of the width

	// Create a vertical flex layout for the content and status bar
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	// Set up the UI
	app.SetRoot(mainFlex, true).SetFocus(tree)

	// Add a global capture for key commands
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		} else if event.Rune() == 'r' {
			// Run pytest for the selected test
			node := tree.GetCurrentNode()
			if node != nil {
				// Get the full test path
				testPath := getTestPath(node)
				if testPath != "" {
					// Update status bar to show we're running the test
					statusBar.SetText(fmt.Sprintf("Running: %s", testPath))

					// Clear the output view
					outputView.Clear()

					// Run pytest in a goroutine
					go func() {
						// Run the test with verbose output
						cmd := exec.Command("pytest", testPath, "-vv")

						// Create a pipe for the command's output
						stdout, err := cmd.StdoutPipe()
						if err != nil {
							app.QueueUpdateDraw(func() {
								fmt.Fprintf(outputView, "[red]Error creating stdout pipe: %v[white]\n", err)
							})
							return
						}

						stderr, err := cmd.StderrPipe()
						if err != nil {
							app.QueueUpdateDraw(func() {
								fmt.Fprintf(outputView, "[red]Error creating stderr pipe: %v[white]\n", err)
							})
							return
						}

						// Start the command
						if err := cmd.Start(); err != nil {
							app.QueueUpdateDraw(func() {
								fmt.Fprintf(outputView, "[red]Error starting pytest: %v[white]\n", err)
							})
							return
						}

						// Read and display stdout in real-time
						go readAndDisplayOutput(stdout, outputView, app)

						// Read and display stderr in real-time
						go readAndDisplayOutput(stderr, outputView, app)

						// Wait for the command to complete
						err = cmd.Wait()

						// Update the status bar with the result
						app.QueueUpdateDraw(func() {
							if err != nil {
								statusBar.SetText(fmt.Sprintf("Test failed: %s", testPath))
							} else {
								statusBar.SetText(fmt.Sprintf("Test passed: %s", testPath))
							}
						})
					}()
				}
			}
			return nil
		}
		return event
	})

	// Run the application
	if err := app.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}

// readAndDisplayOutput reads from a reader and displays the output in the TextView
func readAndDisplayOutput(reader io.Reader, view *tview.TextView, app *tview.Application) {
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			output := buffer[:n]
			app.QueueUpdateDraw(func() {
				// Process ANSI color codes or just display the text
				// For simplicity, we're just displaying the raw text here
				fmt.Fprint(view, string(output))
			})
		}
		if err != nil {
			break
		}
	}
}

// getTestPath returns the full test path for a node
func getTestPath(node *tview.TreeNode) string {
	if node == nil {
		return ""
	}

	// If this is the root node, return empty
	if node.GetText() == "Tests" {
		return ""
	}

	// Get the reference data which contains the full path
	ref := node.GetReference()
	if ref == nil {
		return ""
	}

	// Convert the reference to a string
	path, ok := ref.(string)
	if !ok {
		return ""
	}

	return path
}

// discoverTests runs pytest --collect-only and parses the output
func discoverTests() ([]string, error) {
	// Use a different pytest command that outputs test IDs directly
	cmd := exec.Command("pytest", "--collect-only", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running pytest: %w", err)
	}

	var tests []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, lines starting with "=", and lines containing "tests collected"
		if line != "" &&
			!strings.HasPrefix(line, "=") &&
			!strings.Contains(line, "tests collected") {
			tests = append(tests, line)
		}
	}

	return tests, nil
}

// addTestsToTree adds the discovered tests to the tree view
func addTestsToTree(root *tview.TreeNode, tests []string) {
	// Map to store module and class nodes
	modules := make(map[string]*tview.TreeNode)
	classes := make(map[string]*tview.TreeNode)

	for _, test := range tests {
		parts := strings.Split(test, "::")

		// Handle module
		moduleName := parts[0]
		moduleNode, ok := modules[moduleName]
		if !ok {
			moduleNode = tview.NewTreeNode(moduleName).
				SetColor(tcell.ColorGreen).
				SetSelectable(true).
				SetReference(moduleName) // Store the module path
			root.AddChild(moduleNode)
			modules[moduleName] = moduleNode
		}

		// Handle class if present
		if len(parts) >= 3 {
			className := parts[1]
			classKey := moduleName + "::" + className
			classNode, ok := classes[classKey]
			if !ok {
				classNode = tview.NewTreeNode(className).
					SetColor(tcell.ColorBlue).
					SetSelectable(true).
					SetReference(moduleName + "::" + className) // Store the class path
				moduleNode.AddChild(classNode)
				classes[classKey] = classNode
			}

			// Add test method
			testName := parts[2]
			testNode := tview.NewTreeNode(testName).
				SetColor(tcell.ColorWhite).
				SetSelectable(true).
				SetReference(test) // Store the full test path
			classNode.AddChild(testNode)
		} else if len(parts) == 2 {
			// Handle function-level test (no class)
			testName := parts[1]
			testNode := tview.NewTreeNode(testName).
				SetColor(tcell.ColorWhite).
				SetSelectable(true).
				SetReference(test) // Store the full test path
			moduleNode.AddChild(testNode)
		}
	}
}
