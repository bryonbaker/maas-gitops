package clusterrolebindings

import (
	"fmt"

	"github.com/bryon/ocp-lister/internal/menu"
	"k8s.io/client-go/kubernetes"
)

// HandleCRUDMenu handles the CRUD menu for cluster role bindings
func HandleCRUDMenu(clientset *kubernetes.Clientset) {
	crudMenu := menu.NewCRUDMenu("Cluster Role Bindings")

	for {
		choice := crudMenu.DisplayAndGetChoice()

		switch choice {
		case "1": // List
			fmt.Println("List cluster role bindings - Not yet implemented")

		case "2": // Get
			name := menu.GetName("Enter cluster role binding name: ")
			fmt.Printf("Get cluster role binding %s - Not yet implemented\n", name)

		case "3": // Create
			name := menu.GetName("Enter cluster role binding name to create: ")
			fmt.Printf("Create cluster role binding %s - Not yet implemented\n", name)

		case "4": // Update
			name := menu.GetName("Enter cluster role binding name to update: ")
			fmt.Printf("Update cluster role binding %s - Not yet implemented\n", name)

		case "5": // Delete
			name := menu.GetName("Enter cluster role binding name to delete: ")
			fmt.Printf("Delete cluster role binding %s - Not yet implemented\n", name)

		case "6": // Add Annotation
			name := menu.GetName("Enter cluster role binding name to annotate: ")
			if name == "" {
				fmt.Println("Cluster role binding name cannot be empty")
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
