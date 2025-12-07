package menu

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Menu represents a menu with options
type Menu struct {
	Title   string
	Options map[string]string
}

// NewMenu creates a new menu
func NewMenu(title string) *Menu {
	return &Menu{
		Title:   title,
		Options: make(map[string]string),
	}
}

// AddOption adds an option to the menu
func (m *Menu) AddOption(key, description string) {
	m.Options[key] = description
}

// Display shows the menu and returns the selected option
func (m *Menu) Display() (string, error) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println(m.Title)
	fmt.Println(strings.Repeat("=", 50))

	// Sort and display options
	keys := []string{}
	for key := range m.Options {
		keys = append(keys, key)
	}

	// Display in order (a, b, c, etc. then X)
	for _, key := range keys {
		if key != "X" && key != "x" {
			fmt.Printf("%s. %s\n", strings.ToUpper(key), m.Options[key])
		}
	}

	// Always show exit option last
	if exitDesc, exists := m.Options["X"]; exists {
		fmt.Printf("X. %s\n", exitDesc)
	} else if exitDesc, exists := m.Options["x"]; exists {
		fmt.Printf("X. %s\n", exitDesc)
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Print("Select an option: ")

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	choice = strings.TrimSpace(strings.ToUpper(choice))

	// Validate choice
	if _, exists := m.Options[choice]; !exists {
		return "", fmt.Errorf("invalid option: %s", choice)
	}

	return choice, nil
}

// DisplayAndGetChoice displays the menu and returns the choice, handling errors
func (m *Menu) DisplayAndGetChoice() string {
	for {
		choice, err := m.Display()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		return choice
	}
}
