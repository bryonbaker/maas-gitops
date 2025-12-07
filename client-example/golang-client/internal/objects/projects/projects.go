package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListProjects retrieves and returns a list of all projects (namespaces) the user has access to
func ListProjects(clientset *kubernetes.Clientset) ([]string, error) {
	ctx := context.Background()

	// List all namespaces (in OpenShift, projects are namespaces)
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Extract namespace names
	projects := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		projects = append(projects, ns.Name)
	}

	return projects, nil
}

// PrintProjects prints the list of projects to stdout
func PrintProjects(projects []string) {
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return
	}

	fmt.Printf("\nFound %d project(s):\n\n", len(projects))
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project)
	}
	fmt.Println()
}

// HandleList handles the list action for projects
func HandleList(clientset *kubernetes.Clientset) error {
	projectList, err := ListProjects(clientset)
	if err != nil {
		return fmt.Errorf("error listing projects: %w", err)
	}
	PrintProjects(projectList)
	return nil
}

// HandleGet handles the get action for a specific project
func HandleGet(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	namespace, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(namespace, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling project to JSON: %w", err)
	}

	fmt.Println("\n" + string(jsonData))
	fmt.Println()

	return nil
}

// HandleCreate handles the create action for projects
func HandleCreate(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	// Validate project name (Kubernetes namespace naming rules)
	if err := validateProjectName(name); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Check if project already exists
	_, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return fmt.Errorf("project '%s' already exists", name)
	}

	// Create the namespace object
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name": name,
			},
		},
	}

	// Create the namespace
	created, err := clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("\n✓ Successfully created project: %s\n", created.Name)
	fmt.Printf("  Status: %s\n", created.Status.Phase)
	fmt.Printf("  Created: %s\n", created.CreationTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Println()

	return nil
}

// validateProjectName validates a project name according to Kubernetes naming rules
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("project name cannot be longer than 63 characters")
	}

	// Kubernetes DNS-1123 subdomain rules
	// Must be lowercase alphanumeric characters or '-' or '.'
	// Must start and end with alphanumeric character
	for i, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.') {
			return fmt.Errorf("project name contains invalid character '%c' at position %d (only lowercase alphanumeric, '-', and '.' are allowed)", r, i)
		}
	}

	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= '0' && name[0] <= '9')) {
		return fmt.Errorf("project name must start with a lowercase alphanumeric character")
	}

	if !((name[len(name)-1] >= 'a' && name[len(name)-1] <= 'z') || (name[len(name)-1] >= '0' && name[len(name)-1] <= '9')) {
		return fmt.Errorf("project name must end with a lowercase alphanumeric character")
	}

	// Check for uppercase letters
	if strings.ToLower(name) != name {
		return fmt.Errorf("project name must be lowercase")
	}

	return nil
}

// HandleUpdate handles the update action for projects (placeholder)
func HandleUpdate(clientset *kubernetes.Clientset, name string) error {
	fmt.Printf("Update project functionality not yet implemented for: %s\n", name)
	return nil
}

// HandleDelete handles the delete action for projects
func HandleDelete(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	// First, verify the project exists
	namespace, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	// Show project details before deletion
	fmt.Printf("\nProject to delete: %s\n", namespace.Name)
	fmt.Printf("Status: %s\n", namespace.Status.Phase)
	fmt.Printf("Created: %s\n", namespace.CreationTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("\n⚠️  WARNING: This will delete the project and all resources within it!")
	fmt.Println("   This action cannot be undone.")
	fmt.Println()

	// Delete the namespace
	err = clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting project: %w", err)
	}

	fmt.Printf("✓ Successfully initiated deletion of project: %s\n", name)
	fmt.Println("  Note: Project deletion is asynchronous and may take some time to complete.")
	fmt.Println()

	return nil
}

// HandleAddAnnotation adds the annotation "bakerapps.net/test": "annotated" to a project
func HandleAddAnnotation(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	// Get the existing namespace
	namespace, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	// Initialize annotations map if nil
	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string)
	}

	// Add the annotation
	namespace.Annotations["bakerapps.net/test"] = "annotated"

	// Update the namespace
	updated, err := clientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating project with annotation: %w", err)
	}

	fmt.Printf("\n✓ Successfully added annotation to project: %s\n", updated.Name)
	fmt.Printf("  Annotation: bakerapps.net/test = annotated\n")
	fmt.Println()

	return nil
}
