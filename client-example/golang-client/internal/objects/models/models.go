package models

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

// getModelClient creates a dynamic client for LLMInferenceService resources
func getModelClient(clientset *kubernetes.Clientset) (dynamic.Interface, error) {
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

// getModelResource returns the GVR for LLMInferenceService resources
func getModelResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "serving.kserve.io",
		Version:  "v1alpha1",
		Resource: "llminferenceservices",
	}
}

// HandleDeploy deploys an LLMInferenceService with the specified name and namespace
// All other fields are set exactly as in the GitHub example
func HandleDeploy(clientset *kubernetes.Clientset, name, namespace string) error {
	ctx := context.Background()

	// Check if namespace exists
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace '%s' does not exist: %w", namespace, err)
	}

	dynamicClient, err := getModelClient(clientset)
	if err != nil {
		return err
	}

	// Check if model already exists
	_, err = dynamicClient.Resource(getModelResource()).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return fmt.Errorf("model '%s' already exists in namespace '%s'", name, namespace)
	}

	// Create the LLMInferenceService object exactly as in the GitHub example
	model := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.kserve.io/v1alpha1",
			"kind":       "LLMInferenceService",
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"alpha.maas.opendatahub.io/tiers": `["redhat-users-tier"]`,
				},
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"model": map[string]interface{}{
					"name": "facebook/opt-125m",
					"uri":  "hf://facebook/opt-125m",
				},
				"replicas": int64(1),
				"router": map[string]interface{}{
					"gateway": map[string]interface{}{
						"refs": []interface{}{
							map[string]interface{}{
								"name":      "maas-default-gateway",
								"namespace": "openshift-ingress",
							},
						},
					},
					"route": map[string]interface{}{},
				},
				"template": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"args": []interface{}{
								"--port",
								"8000",
								"--model",
								"facebook/opt-125m",
								"--mode",
								"random",
								"--ssl-certfile",
								"/var/run/kserve/tls/tls.crt",
								"--ssl-keyfile",
								"/var/run/kserve/tls/tls.key",
							},
							"command": []interface{}{
								"/app/llm-d-inference-sim",
							},
							"env": []interface{}{
								map[string]interface{}{
									"name": "POD_NAME",
									"valueFrom": map[string]interface{}{
										"fieldRef": map[string]interface{}{
											"apiVersion": "v1",
											"fieldPath":  "metadata.name",
										},
									},
								},
								map[string]interface{}{
									"name": "POD_NAMESPACE",
									"valueFrom": map[string]interface{}{
										"fieldRef": map[string]interface{}{
											"apiVersion": "v1",
											"fieldPath":  "metadata.namespace",
										},
									},
								},
							},
							"image":           "ghcr.io/llm-d/llm-d-inference-sim:v0.5.1",
							"imagePullPolicy": "Always",
							"livenessProbe": map[string]interface{}{
								"httpGet": map[string]interface{}{
									"path":   "/health",
									"port":   "https",
									"scheme": "HTTPS",
								},
							},
							"name": "main",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": int64(8000),
									"name":          "https",
									"protocol":      "TCP",
								},
							},
							"readinessProbe": map[string]interface{}{
								"httpGet": map[string]interface{}{
									"path":   "/ready",
									"port":   "https",
									"scheme": "HTTPS",
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the model
	created, err := dynamicClient.Resource(getModelResource()).Namespace(namespace).Create(ctx, model, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to deploy model: %w", err)
	}

	createdName, _, _ := unstructured.NestedString(created.Object, "metadata", "name")
	fmt.Printf("\n✓ Successfully deployed model: %s\n", createdName)
	fmt.Printf("  Namespace: %s\n", namespace)
	fmt.Printf("  API Version: serving.kserve.io/v1alpha1\n")
	fmt.Println()

	return nil
}

// HandleUndeploy removes an LLMInferenceService
func HandleUndeploy(clientset *kubernetes.Clientset, name, namespace string) error {
	ctx := context.Background()

	dynamicClient, err := getModelClient(clientset)
	if err != nil {
		return err
	}

	// Get model first to verify it exists
	model, err := dynamicClient.Resource(getModelResource()).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting model: %w", err)
	}

	modelName, _, _ := unstructured.NestedString(model.Object, "metadata", "name")

	// Show model details before deletion
	fmt.Printf("\nModel to undeploy: %s\n", modelName)
	fmt.Printf("Namespace: %s\n", namespace)
	fmt.Println("\n⚠️  WARNING: This will undeploy the model!")
	fmt.Println("   This action cannot be undone.")
	fmt.Println()

	// Delete the model
	err = dynamicClient.Resource(getModelResource()).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error undeploying model: %w", err)
	}

	fmt.Printf("✓ Successfully undeployed model: %s\n", name)
	fmt.Println()

	return nil
}

// HandleList lists all LLMInferenceService models in the specified namespace
func HandleList(clientset *kubernetes.Clientset, namespace string) error {
	ctx := context.Background()

	dynamicClient, err := getModelClient(clientset)
	if err != nil {
		return err
	}

	// List models in specified namespace
	modelList, err := dynamicClient.Resource(getModelResource()).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(modelList.Items) == 0 {
		fmt.Printf("\nNo models found in namespace '%s'.\n", namespace)
		fmt.Println()
		return nil
	}

	fmt.Printf("\nFound %d model(s) in namespace '%s':\n\n", len(modelList.Items), namespace)
	for i, model := range modelList.Items {
		name, _, _ := unstructured.NestedString(model.Object, "metadata", "name")
		fmt.Printf("%d. %s\n", i+1, name)
	}
	fmt.Println()

	return nil
}

// HandleGet retrieves and displays a specific model as JSON
func HandleGet(clientset *kubernetes.Clientset, name, namespace string) error {
	ctx := context.Background()

	dynamicClient, err := getModelClient(clientset)
	if err != nil {
		return err
	}

	// Get model
	model, err := dynamicClient.Resource(getModelResource()).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting model: %w", err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(model.Object, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling model to JSON: %w", err)
	}

	fmt.Println("\n" + string(jsonData))
	fmt.Println()

	return nil
}
