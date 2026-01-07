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

// HandleAllow handles the /allow endpoint
// This endpoint is ONLY called for inference requests because:
// - Route-level AuthPolicies handle non-inference paths (/v1/models, /maas-api/*, /oauth/*)
// - Route-level policies take precedence over gateway-level policies
// - This gateway-level policy only processes inference requests: /{namespace}/{model}/v1/*
//
// Headers received from AuthPolicy:
//   - X-Original-Path: The original request path (for logging only)
//   - X-Username: The authenticated username (for logging and future credit lookup)
//   - X-Method: The HTTP method (for logging only)
func (h *CreditCheckHandler) HandleAllow(c *gin.Context) {
	startTime := time.Now()

	// Extract headers for logging and future credit lookup
	originalPath := c.GetHeader("X-Original-Path")
	username := c.GetHeader("X-Username")
	method := c.GetHeader("X-Method")

	// Perform credit check
	// Phase 1: Use ALLOW_VALUE environment variable (true/false)
	// Phase 2+: Implement actual credit check logic:
	//   - Query database for user's credit balance
	//   - Or check ConfigMap for user quotas
	//   - Or call external billing API
	allow := h.allowValue

	duration := time.Since(startTime).Milliseconds()

	// Structured logging - log every inference request for observability
	log.Printf("[CREDIT-CHECK] path=%s user=%s method=%s allow=%v duration_ms=%d",
		originalPath, username, method, allow, duration)

	// Always return 200 OK with credit decision
	// This prevents AuthPolicy reconciliation loops
	c.JSON(http.StatusOK, AllowResponse{Allow: allow})
}

// HandleHealth handles the /health endpoint
func HandleHealth(c *gin.Context) {
	response := HealthResponse{
		Status: "healthy",
	}

	c.JSON(http.StatusOK, response)
}

