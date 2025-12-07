package main

import (
	"fmt"
	"os"

	"github.com/bryon/ocp-lister/internal/auth"
	"github.com/bryon/ocp-lister/internal/client"
	"github.com/bryon/ocp-lister/internal/menu"
	"github.com/bryon/ocp-lister/internal/objects/clusterrolebindings"
	"github.com/bryon/ocp-lister/internal/objects/groups"
	"github.com/bryon/ocp-lister/internal/objects/models"
	"github.com/bryon/ocp-lister/internal/objects/projects"
	"github.com/bryon/ocp-lister/internal/objects/users"
)

func main() {
	// Load authentication configuration from environment variables
	authConfig, err := auth.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connecting to OpenShift cluster at %s...\n", authConfig.Server)

	// Create Kubernetes client
	clientset, err := client.CreateClient(
		authConfig.Server,
		authConfig.Username,
		authConfig.Password,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully authenticated!")

	// Create main menu
	mainMenu := menu.NewMenu("OpenShift Kubernetes Object Manager")
	mainMenu.AddOption("A", "Projects")
	mainMenu.AddOption("B", "Groups")
	mainMenu.AddOption("C", "Users")
	mainMenu.AddOption("D", "Cluster Role Bindings")
	mainMenu.AddOption("E", "Model")
	mainMenu.AddOption("X", "Exit")

	// Main menu loop
	for {
		choice := mainMenu.DisplayAndGetChoice()

		switch choice {
		case "A":
			projects.HandleCRUDMenu(clientset)
		case "B":
			groups.HandleCRUDMenu(clientset)
		case "C":
			users.HandleCRUDMenu(clientset)
		case "D":
			clusterrolebindings.HandleCRUDMenu(clientset)
		case "E":
			models.HandleModelMenu(clientset)
		case "X":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Printf("Unknown option: %s\n", choice)
		}
	}
}
