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

package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AllowResponse represents the credit check response
type AllowResponse struct {
	Allow bool `json:"allow"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// CreditCheckHandler handles credit check requests
type CreditCheckHandler struct {
	allowValue bool
}

// NewCreditCheckHandler creates a new credit check handler
func NewCreditCheckHandler(allowValue bool) *CreditCheckHandler {
	return &CreditCheckHandler{
		allowValue: allowValue,
	}
}

// HandleAllow handles the /allow endpoint with request path filtering
// This endpoint is called by Authorino via AuthPolicy metadata fetch
// Headers expected from AuthPolicy:
//   - X-Original-Path: The original request path
//   - X-Username: The authenticated username
//   - X-Method: The HTTP method
func (h *CreditCheckHandler) HandleAllow(c *gin.Context) {
	startTime := time.Now()

	// Extract request context from headers (passed by AuthPolicy)
	originalPath := c.GetHeader("X-Original-Path")
	username := c.GetHeader("X-Username")
	method := c.GetHeader("X-Method")

	// Determine request type and make decision
	requestType, allow := h.evaluateRequest(originalPath, method, username)

	duration := time.Since(startTime).Milliseconds()

	// Structured logging - log every request
	log.Printf("[CREDIT-CHECK] path=%s user=%s method=%s type=%s allow=%v duration_ms=%d",
		originalPath, username, method, requestType, allow, duration)

	// Always return 200 OK with allow decision
	// This prevents AuthPolicy reconciliation loops
	c.JSON(http.StatusOK, AllowResponse{Allow: allow})
}

// evaluateRequest determines if the request should be allowed
// Returns: (requestType, allowDecision)
func (h *CreditCheckHandler) evaluateRequest(path, method, username string) (string, bool) {
	// Fast path: Non-inference requests always allowed (no credit check)
	if isNonInferencePath(path) {
		return "non-inference", true
	}

	// Inference path: perform credit check
	if isInferencePath(path) {
		// For MVP: use ALLOW_VALUE environment variable
		// For production: implement actual credit check logic here
		// e.g., check database, ConfigMap, external API, etc.
		return "inference", h.allowValue
	}

	// Unknown paths: deny by default (fail-closed)
	log.Printf("[CREDIT-CHECK] WARN: unknown path pattern: %s", path)
	return "unknown", false
}

// isNonInferencePath checks if the path is a non-inference request
// These paths bypass credit checking entirely
func isNonInferencePath(path string) bool {
	// API endpoints - always bypass credit check
	if strings.HasPrefix(path, "/maas-api") {
		return true
	}

	// Model listing endpoint (OpenAI-compatible API)
	// GET /v1/models - lists available models
	if strings.HasPrefix(path, "/v1/models") {
		return true
	}

	// OAuth endpoints - always bypass credit check
	if strings.HasPrefix(path, "/oauth") {
		return true
	}

	// Health check endpoints
	if strings.HasPrefix(path, "/health") {
		return true
	}

	// Root path
	if path == "/" {
		return true
	}

	return false
}

// isInferencePath checks if the path is an inference request
// Pattern: /{namespace}/{model}/v1/{endpoint}
//
// The presence of "/v1/" as the 3rd segment indicates this is an
// OpenAI-compatible API endpoint for model inference.
//
// Examples:
//   - /llm/facebook-opt-125m-simulated/v1/completions
//   - /acme-inc-models/acme-inc-model-1/v1/chat/completions
//   - /serverless-models/my-model-name/v1/completions
//
// The "v1" refers to the OpenAI API version that MaaS supports.
// This pattern works for any namespace without hardcoding namespace names.
func isInferencePath(path string) bool {
	// Fast path: if it's explicitly non-inference, skip
	if isNonInferencePath(path) {
		return false
	}

	// Split path into segments
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return false
	}

	parts := strings.Split(trimmed, "/")

	// Inference paths must have at least 3 segments: {namespace}/{model}/v1
	// And the 3rd segment (index 2) must be "v1" (OpenAI API version)
	//
	// Pattern breakdown:
	//   parts[0] = namespace (e.g., "llm", "acme-inc-models")
	//   parts[1] = model name (e.g., "facebook-opt-125m-simulated")
	//   parts[2] = API version (must be "v1")
	//   parts[3+] = endpoint (e.g., "completions", "chat/completions")
	if len(parts) >= 3 && parts[2] == "v1" {
		return true
	}

	return false
}

// HandleHealth handles the /health endpoint
func HandleHealth(c *gin.Context) {
	response := HealthResponse{
		Status: "healthy",
	}

	c.JSON(http.StatusOK, response)
}

