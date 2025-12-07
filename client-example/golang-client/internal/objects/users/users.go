package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bryon/ocp-lister/internal/auth"
	"github.com/bryon/ocp-lister/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// getUserClient creates a dynamic client for User resources
func getUserClient(clientset *kubernetes.Clientset) (dynamic.Interface, error) {
	// Get auth config to retrieve server, username, password
	authConfig, err := auth.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load auth config: %w", err)
	}

	// Get REST config
	config, err := client.GetRESTConfig(authConfig.Server, authConfig.Username, authConfig.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return dynamicClient, nil
}

// getUserResource returns the GVR for User resources
func getUserResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "user.openshift.io",
		Version:  "v1",
		Resource: "users",
	}
}

// ListUsers retrieves and returns a list of all users
func ListUsers(clientset *kubernetes.Clientset) ([]string, error) {
	ctx := context.Background()

	dynamicClient, err := getUserClient(clientset)
	if err != nil {
		return nil, err
	}

	// List users
	userList, err := dynamicClient.Resource(getUserResource()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Extract user names
	users := make([]string, 0, len(userList.Items))
	for _, user := range userList.Items {
		if name, found, _ := unstructured.NestedString(user.Object, "metadata", "name"); found {
			users = append(users, name)
		}
	}

	return users, nil
}

// PrintUsers prints the list of users to stdout
func PrintUsers(users []string) {
	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	fmt.Printf("\nFound %d user(s):\n\n", len(users))
	for i, user := range users {
		fmt.Printf("%d. %s\n", i+1, user)
	}
	fmt.Println()
}

// HandleList handles the list action for users
func HandleList(clientset *kubernetes.Clientset) error {
	userList, err := ListUsers(clientset)
	if err != nil {
		return fmt.Errorf("error listing users: %w", err)
	}
	PrintUsers(userList)
	return nil
}

// HandleGet handles the get action for a specific user
func HandleGet(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	dynamicClient, err := getUserClient(clientset)
	if err != nil {
		return err
	}

	// Get user
	user, err := dynamicClient.Resource(getUserResource()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(user.Object, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling user to JSON: %w", err)
	}

	fmt.Println("\n" + string(jsonData))
	fmt.Println()

	return nil
}

// HandleCreate handles the create action for users
func HandleCreate(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	dynamicClient, err := getUserClient(clientset)
	if err != nil {
		return err
	}

	// Check if user already exists
	_, err = dynamicClient.Resource(getUserResource()).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return fmt.Errorf("user '%s' already exists", name)
	}

	// Create the user object
	user := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "user.openshift.io/v1",
			"kind":       "User",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}

	// Create the user
	created, err := dynamicClient.Resource(getUserResource()).Create(ctx, user, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	createdName, _, _ := unstructured.NestedString(created.Object, "metadata", "name")
	fmt.Printf("\n✓ Successfully created user: %s\n", createdName)
	fmt.Println()

	return nil
}

// HandleUpdate handles the update action for users (placeholder)
func HandleUpdate(clientset *kubernetes.Clientset, name string) error {
	fmt.Printf("Update user functionality not yet implemented for: %s\n", name)
	return nil
}

// HandleDelete handles the delete action for users
func HandleDelete(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	dynamicClient, err := getUserClient(clientset)
	if err != nil {
		return err
	}

	// Get user first to show details
	user, err := dynamicClient.Resource(getUserResource()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	userName, _, _ := unstructured.NestedString(user.Object, "metadata", "name")
	created, _, _ := unstructured.NestedString(user.Object, "metadata", "creationTimestamp")

	// Show user details before deletion
	fmt.Printf("\nUser to delete: %s\n", userName)
	if created != "" {
		fmt.Printf("Created: %s\n", created)
	}
	fmt.Println("\n⚠️  WARNING: This will delete the user!")
	fmt.Println("   This action cannot be undone.")
	fmt.Println()

	// Delete the user
	err = dynamicClient.Resource(getUserResource()).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	fmt.Printf("✓ Successfully deleted user: %s\n", name)
	fmt.Println()

	return nil
}

// HandleAddAnnotation adds the annotation "bakerapps.net/test": "annotated" to a user
func HandleAddAnnotation(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	dynamicClient, err := getUserClient(clientset)
	if err != nil {
		return err
	}

	// Get the existing user
	user, err := dynamicClient.Resource(getUserResource()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Get or create annotations map
	annotations, found, err := unstructured.NestedStringMap(user.Object, "metadata", "annotations")
	if err != nil {
		return fmt.Errorf("error getting annotations: %w", err)
	}
	if !found || annotations == nil {
		annotations = make(map[string]string)
	}

	// Add the annotation
	annotations["bakerapps.net/test"] = "annotated"

	// Set annotations back
	if err := unstructured.SetNestedStringMap(user.Object, annotations, "metadata", "annotations"); err != nil {
		return fmt.Errorf("error setting annotations: %w", err)
	}

	// Update the user
	updated, err := dynamicClient.Resource(getUserResource()).Update(ctx, user, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating user with annotation: %w", err)
	}

	updatedName, _, _ := unstructured.NestedString(updated.Object, "metadata", "name")
	fmt.Printf("\n✓ Successfully added annotation to user: %s\n", updatedName)
	fmt.Printf("  Annotation: bakerapps.net/test = annotated\n")
	fmt.Println()

	return nil
}
