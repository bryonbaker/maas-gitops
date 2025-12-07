package menu

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CRUDMenu represents a CRUD menu for a Kubernetes object
type CRUDMenu struct {
	ObjectType string
}

// NewCRUDMenu creates a new CRUD menu
func NewCRUDMenu(objectType string) *CRUDMenu {
	return &CRUDMenu{
		ObjectType: objectType,
	}
}

// Display shows the CRUD menu and returns the selected action
func (c *CRUDMenu) Display() (string, error) {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Printf("%s Management\n", titleCase(c.ObjectType))
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("1. List (Read)")
	fmt.Println("2. Get (Read by name)")
	fmt.Println("3. Create")
	fmt.Println("4. Update")
	fmt.Println("5. Delete")
	fmt.Println("6. Add Annotation")
	fmt.Println("B. Back to main menu")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Print("Select an action: ")

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	choice = strings.TrimSpace(strings.ToUpper(choice))

	validChoices := map[string]bool{
		"1": true, "2": true, "3": true, "4": true, "5": true, "6": true, "B": true,
	}

	if !validChoices[choice] {
		return "", fmt.Errorf("invalid option: %s", choice)
	}

	return choice, nil
}

// DisplayAndGetChoice displays the CRUD menu and returns the choice, handling errors
func (c *CRUDMenu) DisplayAndGetChoice() string {
	for {
		choice, err := c.Display()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		return choice
	}
}

// titleCase converts a string to title case (simple implementation)
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// GetName prompts for a resource name
func GetName(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	return strings.TrimSpace(name)
}

// GetConfirmation prompts for yes/no confirmation
func GetConfirmation(prompt string) bool {
	fmt.Print(prompt + " (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
