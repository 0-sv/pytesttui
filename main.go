package main

import (
	"fmt"
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

	// Discover pytest tests
	tests, err := discoverTests()
	if err != nil {
		fmt.Printf("Error discovering tests: %v\n", err)
		os.Exit(1)
	}

	// Add tests to the tree
	addTestsToTree(root, tests)

	// Set up the UI
	app.SetRoot(tree, true).SetFocus(tree)

	// Run the application
	if err := app.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
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
		if line != "" && !strings.HasPrefix(line, "=") {
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
				SetSelectable(true)
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
					SetSelectable(true)
				moduleNode.AddChild(classNode)
				classes[classKey] = classNode
			}

			// Add test method
			testName := parts[2]
			testNode := tview.NewTreeNode(testName).
				SetColor(tcell.ColorWhite).
				SetSelectable(true)
			classNode.AddChild(testNode)
		} else if len(parts) == 2 {
			// Handle function-level test (no class)
			testName := parts[1]
			testNode := tview.NewTreeNode(testName).
				SetColor(tcell.ColorWhite).
				SetSelectable(true)
			moduleNode.AddChild(testNode)
		}
	}
}
