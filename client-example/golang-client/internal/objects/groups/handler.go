package groups

import (
	"fmt"

	"github.com/bryon/ocp-lister/internal/menu"
	"k8s.io/client-go/kubernetes"
)

// HandleCRUDMenu handles the CRUD menu for groups
func HandleCRUDMenu(clientset *kubernetes.Clientset) {
	crudMenu := menu.NewCRUDMenu("Groups")

	for {
		choice := crudMenu.DisplayAndGetChoice()

		switch choice {
		case "1": // List
			fmt.Println("List groups - Not yet implemented")

		case "2": // Get
			name := menu.GetName("Enter group name: ")
			fmt.Printf("Get group %s - Not yet implemented\n", name)

		case "3": // Create
			name := menu.GetName("Enter group name to create: ")
			fmt.Printf("Create group %s - Not yet implemented\n", name)

		case "4": // Update
			name := menu.GetName("Enter group name to update: ")
			fmt.Printf("Update group %s - Not yet implemented\n", name)

		case "5": // Delete
			name := menu.GetName("Enter group name to delete: ")
			fmt.Printf("Delete group %s - Not yet implemented\n", name)

		case "6": // Add Annotation
			name := menu.GetName("Enter group name to annotate: ")
			if name == "" {
				fmt.Println("Group name cannot be empty")
				continue
			}
			fmt.Printf("Add annotation to group %s - Not yet implemented (requires OpenShift client)\n", name)

		case "B": // Back
			return
		}
	}
}
