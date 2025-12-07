package models

import (
	"fmt"

	"github.com/bryon/ocp-lister/internal/menu"
	"k8s.io/client-go/kubernetes"
)

// HandleModelMenu handles the model menu with Deploy and Undeploy options
func HandleModelMenu(clientset *kubernetes.Clientset) {
	modelMenu := menu.NewMenu("Model Management")
	modelMenu.AddOption("1", "Deploy")
	modelMenu.AddOption("2", "Undeploy")
	modelMenu.AddOption("3", "List")
	modelMenu.AddOption("4", "Get")
	modelMenu.AddOption("B", "Back to main menu")

	for {
		choice := modelMenu.DisplayAndGetChoice()

		switch choice {
		case "1": // Deploy
			name := menu.GetName("Enter model name to deploy: ")
			if name == "" {
				fmt.Println("Model name cannot be empty")
				continue
			}
			namespace := menu.GetName("Enter namespace: ")
			if namespace == "" {
				fmt.Println("Namespace cannot be empty")
				continue
			}
			if err := HandleDeploy(clientset, name, namespace); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "2": // Undeploy
			name := menu.GetName("Enter model name to undeploy: ")
			if name == "" {
				fmt.Println("Model name cannot be empty")
				continue
			}
			namespace := menu.GetName("Enter namespace: ")
			if namespace == "" {
				fmt.Println("Namespace cannot be empty")
				continue
			}
			// Get confirmation before undeploying
			if !menu.GetConfirmation(fmt.Sprintf("Are you sure you want to undeploy model '%s' in namespace '%s'", name, namespace)) {
				fmt.Println("Undeploy cancelled.")
				continue
			}
			if err := HandleUndeploy(clientset, name, namespace); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "3": // List
			namespace := menu.GetName("Enter namespace (or press Enter for 'llm'): ")
			if namespace == "" {
				namespace = "llm"
			}
			if err := HandleList(clientset, namespace); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "4": // Get
			name := menu.GetName("Enter model name: ")
			if name == "" {
				fmt.Println("Model name cannot be empty")
				continue
			}
			namespace := menu.GetName("Enter namespace (or press Enter for 'llm'): ")
			if namespace == "" {
				namespace = "llm"
			}
			if err := HandleGet(clientset, name, namespace); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "B": // Back
			return
		}
	}
}
