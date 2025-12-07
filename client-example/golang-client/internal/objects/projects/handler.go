package projects

import (
	"fmt"

	"github.com/bryon/ocp-lister/internal/menu"
	"k8s.io/client-go/kubernetes"
)

// HandleCRUDMenu handles the CRUD menu for projects
func HandleCRUDMenu(clientset *kubernetes.Clientset) {
	crudMenu := menu.NewCRUDMenu("Projects")

	for {
		choice := crudMenu.DisplayAndGetChoice()

		switch choice {
		case "1": // List
			if err := HandleList(clientset); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "2": // Get
			name := menu.GetName("Enter project name: ")
			if name == "" {
				fmt.Println("Project name cannot be empty")
				continue
			}
			if err := HandleGet(clientset, name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "3": // Create
			name := menu.GetName("Enter project name to create: ")
			if name == "" {
				fmt.Println("Project name cannot be empty")
				continue
			}
			if err := HandleCreate(clientset, name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "4": // Update
			name := menu.GetName("Enter project name to update: ")
			if name == "" {
				fmt.Println("Project name cannot be empty")
				continue
			}
			if err := HandleUpdate(clientset, name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "5": // Delete
			name := menu.GetName("Enter project name to delete: ")
			if name == "" {
				fmt.Println("Project name cannot be empty")
				continue
			}
			// Get confirmation before deleting
			if !menu.GetConfirmation(fmt.Sprintf("Are you sure you want to delete project '%s'", name)) {
				fmt.Println("Deletion cancelled.")
				continue
			}
			if err := HandleDelete(clientset, name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "6": // Add Annotation
			name := menu.GetName("Enter project name to annotate: ")
			if name == "" {
				fmt.Println("Project name cannot be empty")
				continue
			}
			if err := HandleAddAnnotation(clientset, name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "B": // Back
			return
		}
	}
}
