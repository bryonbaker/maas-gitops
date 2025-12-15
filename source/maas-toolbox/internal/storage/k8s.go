// Copyright 2025 Bryon Baker
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"maas-toolbox/internal/models"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sTierStorage implements TierStorage using Kubernetes ConfigMap
type K8sTierStorage struct {
	Client    kubernetes.Interface
	Namespace string
	ConfigMap string
}

// NewK8sTierStorage creates a new K8sTierStorage instance
func NewK8sTierStorage(client kubernetes.Interface, namespace, configMap string) *K8sTierStorage {
	return &K8sTierStorage{
		Client:    client,
		Namespace: namespace,
		ConfigMap: configMap,
	}
}

// Load retrieves the tier configuration from Kubernetes ConfigMap
func (k *K8sTierStorage) Load() (*models.TierConfig, error) {
	ctx := context.Background()
	log.Printf("Loading ConfigMap: namespace=%s, name=%s", k.Namespace, k.ConfigMap)

	// Get ConfigMap from Kubernetes API
	cm, err := k.Client.CoreV1().ConfigMaps(k.Namespace).Get(ctx, k.ConfigMap, metav1.GetOptions{})
	if err != nil {
		// If ConfigMap doesn't exist, return empty config
		if errors.IsNotFound(err) {
			log.Printf("ConfigMap %s/%s not found, returning empty config", k.Namespace, k.ConfigMap)
			return &models.TierConfig{Tiers: []models.Tier{}}, nil
		}
		log.Printf("Error getting ConfigMap %s/%s: %v", k.Namespace, k.ConfigMap, err)
		return nil, fmt.Errorf("failed to get ConfigMap %s/%s: %w", k.Namespace, k.ConfigMap, err)
	}

	log.Printf("ConfigMap retrieved successfully")

	// Extract the "tiers" field from data
	tiersYAML, exists := cm.Data["tiers"]
	if !exists {
		log.Printf("ConfigMap %s/%s does not have 'tiers' key. Available keys: %v", k.Namespace, k.ConfigMap, getMapKeys(cm.Data))
		return &models.TierConfig{Tiers: []models.Tier{}}, nil
	}
	if tiersYAML == "" || tiersYAML == "[]" {
		log.Printf("ConfigMap %s/%s has empty 'tiers' field", k.Namespace, k.ConfigMap)
		return &models.TierConfig{Tiers: []models.Tier{}}, nil
	}

	log.Printf("Parsing tiers YAML (length: %d chars)", len(tiersYAML))

	// Parse the tiers YAML string
	var tiers []models.Tier
	if err := yaml.Unmarshal([]byte(tiersYAML), &tiers); err != nil {
		log.Printf("Failed to parse tiers YAML: %v", err)
		return nil, fmt.Errorf("failed to parse tiers YAML: %w", err)
	}

	log.Printf("Successfully loaded %d tiers from ConfigMap", len(tiers))
	return &models.TierConfig{Tiers: tiers}, nil
}

// Helper function to get keys from a map for logging
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Save persists the tier configuration to Kubernetes ConfigMap
func (k *K8sTierStorage) Save(config *models.TierConfig) error {
	ctx := context.Background()

	// Marshal tiers to YAML string with 2-space indentation
	var tiersBuffer bytes.Buffer
	tiersEncoder := yaml.NewEncoder(&tiersBuffer)
	tiersEncoder.SetIndent(2)
	if err := tiersEncoder.Encode(config.Tiers); err != nil {
		return fmt.Errorf("failed to marshal tiers: %w", err)
	}
	tiersEncoder.Close()

	// Remove document separator and trailing newline if present
	tiersYAML := tiersBuffer.String()
	tiersYAML = strings.TrimPrefix(tiersYAML, "---\n")
	tiersYAML = strings.TrimSuffix(tiersYAML, "\n")

	// Try to get existing ConfigMap
	cm, err := k.Client.CoreV1().ConfigMaps(k.Namespace).Get(ctx, k.ConfigMap, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap doesn't exist, create it
			newCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      k.ConfigMap,
					Namespace: k.Namespace,
					Labels: map[string]string{
						"app": "tier-to-group-admin",
					},
				},
				Data: map[string]string{
					"tiers": tiersYAML,
				},
			}

			_, err := k.Client.CoreV1().ConfigMaps(k.Namespace).Create(ctx, newCM, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create ConfigMap: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	// Update existing ConfigMap
	cm.Data["tiers"] = tiersYAML
	_, err = k.Client.CoreV1().ConfigMaps(k.Namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}

	return nil
}

// getRESTConfig creates a REST config for accessing OpenShift resources
// This uses the same logic as NewKubernetesClient to get the config
func getRESTConfig() (*rest.Config, error) {
	// Try in-cluster config first (when running in pod)
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file (for local development)
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
		}
	}
	return config, nil
}

// GroupExists checks if a Group exists in the OpenShift cluster
// Groups are cluster-scoped resources in the user.openshift.io/v1 API group
func (k *K8sTierStorage) GroupExists(groupName string) (bool, error) {
	ctx := context.Background()

	// Get REST config
	config, err := getRESTConfig()
	if err != nil {
		return false, fmt.Errorf("failed to get REST config: %w", err)
	}

	// Create dynamic client for OpenShift resources
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return false, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Define Group resource
	groupResource := schema.GroupVersionResource{
		Group:    "user.openshift.io",
		Version:  "v1",
		Resource: "groups",
	}

	// Try to get the group
	_, err = dynamicClient.Resource(groupResource).Get(ctx, groupName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Group %s not found in cluster", groupName)
			return false, nil
		}
		// For other errors (permission denied, etc.), return the error
		log.Printf("Error checking if group %s exists: %v", groupName, err)
		return false, fmt.Errorf("failed to check if group exists: %w", err)
	}

	log.Printf("Group %s exists in cluster", groupName)
	return true, nil
}
