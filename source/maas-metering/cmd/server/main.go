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
	"maas-metering/internal/api"
	"os"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Get environment variables
	allowValue := os.Getenv("ALLOW_VALUE")
	if allowValue == "" {
		allowValue = "true" // Default to allowing requests for Phase 1
	}

	log.Printf("Starting MaaS Metering Service - Credit Check API")
	log.Printf("Port: %s", *port)
	log.Printf("Allow Value: %s", allowValue)
	log.Printf("Phase: 1 (Minimal Viable Integration - Hello World)")

	// Setup router
	router := api.SetupRouter(allowValue)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Server listening on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET /allow  - Credit check endpoint (returns allow decision)")
	log.Printf("  GET /health - Health check endpoint")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

