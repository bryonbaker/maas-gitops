package clusterrolebindings

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HandleAddAnnotation adds the annotation "bakerapps.net/test": "annotated" to a cluster role binding
func HandleAddAnnotation(clientset *kubernetes.Clientset, name string) error {
	ctx := context.Background()

	// Get the existing cluster role binding
	crb, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting cluster role binding: %w", err)
	}

	// Initialize annotations map if nil
	if crb.Annotations == nil {
		crb.Annotations = make(map[string]string)
	}

	// Add the annotation
	crb.Annotations["bakerapps.net/test"] = "annotated"

	// Update the cluster role binding
	updated, err := clientset.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating cluster role binding with annotation: %w", err)
	}

	fmt.Printf("\n✓ Successfully added annotation to cluster role binding: %s\n", updated.Name)
	fmt.Printf("  Annotation: bakerapps.net/test = annotated\n")
	fmt.Println()

	return nil
}
