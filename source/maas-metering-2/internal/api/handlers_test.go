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
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHandleAllow_WithCredits tests the credit check when user has credits (ALLOW_VALUE=true)
func TestHandleAllow_WithCredits(t *testing.T) {
	// Create handler with ALLOW_VALUE=true (user has credits)
	handler := NewCreditCheckHandler(true)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/allow", nil)

	// Set headers as AuthPolicy would
	c.Request.Header.Set("X-Original-Path", "/llm/test-model/v1/completions")
	c.Request.Header.Set("X-Username", "testuser")
	c.Request.Header.Set("X-Method", "POST")

	// Execute handler
	handler.HandleAllow(c)

	// Verify response
	assert.Equal(t, 200, w.Code)

	var response AllowResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Allow, "Should allow when user has credits")
}

// TestHandleAllow_WithoutCredits tests the credit check when user has no credits (ALLOW_VALUE=false)
func TestHandleAllow_WithoutCredits(t *testing.T) {
	// Create handler with ALLOW_VALUE=false (user has no credits)
	handler := NewCreditCheckHandler(false)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/allow", nil)

	// Set headers as AuthPolicy would
	c.Request.Header.Set("X-Original-Path", "/llm/test-model/v1/completions")
	c.Request.Header.Set("X-Username", "testuser")
	c.Request.Header.Set("X-Method", "POST")

	// Execute handler
	handler.HandleAllow(c)

	// Verify response
	assert.Equal(t, 200, w.Code)

	var response AllowResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Allow, "Should deny when user has no credits")
}

// TestHandleAllow_MissingHeaders tests that the service works even without headers
func TestHandleAllow_MissingHeaders(t *testing.T) {
	// Create handler with ALLOW_VALUE=true
	handler := NewCreditCheckHandler(true)

	// Create test context without headers
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/allow", nil)

	// Execute handler
	handler.HandleAllow(c)

	// Verify response - should still work without headers
	assert.Equal(t, 200, w.Code)

	var response AllowResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Allow, "Should still work without headers")
}

// TestHandleHealth tests the health check endpoint
func TestHandleHealth(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Execute handler
	HandleHealth(c)

	// Verify response
	assert.Equal(t, 200, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
}
