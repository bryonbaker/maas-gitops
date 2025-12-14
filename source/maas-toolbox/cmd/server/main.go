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

package main

import (
	"flag"
	"fmt"
	"log"
	"maas-toolbox/docs"
	"maas-toolbox/internal/api"
	"maas-toolbox/internal/service"
	"maas-toolbox/internal/storage"
	"os"
)

// @title           Open Data Hub MaaS Toolbox API
// @version         1.0
// @description     REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project.
// @description     This API provides CRUD operations for managing tiers that map Kubernetes groups to user-defined subscription tiers.

// @contact.name   Bryon Baker
// @contact.url    https://github.com/bryonbaker

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   https

func init() {
	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "Open Data Hub Maas Toolbox API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project."
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Set Swagger host from environment variable or default to localhost
	// In OpenShift, set ROUTE_HOST environment variable to the route hostname
	swaggerHost := os.Getenv("ROUTE_HOST")
	if swaggerHost == "" {
		swaggerHost = "localhost:8080"
	}
	docs.SwaggerInfo.Host = swaggerHost
}

// @title           Open Data Hub MaaS Toolbox API
// @version         1.0
// @description     REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project.
// @description     This API provides CRUD operations for managing tiers that map Kubernetes groups to user-defined subscription tiers.

// @contact.name   Bryon Baker
// @contact.url    https://github.com/bryonbaker

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   https

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Get environment variables for Kubernetes configuration
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "maas-api"
	}

	configMapName := os.Getenv("CONFIGMAP_NAME")
	if configMapName == "" {
		configMapName = "tier-to-group-mapping"
	}

	// Initialize Kubernetes client
	k8sClient, err := storage.NewKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
		os.Exit(1)
	}

	// Create Kubernetes storage
	tierStorage := storage.NewK8sTierStorage(k8sClient, namespace, configMapName)
	log.Printf("Using Kubernetes ConfigMap storage")
	log.Printf("Namespace: %s", namespace)
	log.Printf("ConfigMap: %s", configMapName)

	// Initialize service
	tierService := service.NewTierService(tierStorage)

	// Setup router
	router := api.SetupRouter(tierService)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting server on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
