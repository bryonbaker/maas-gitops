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
	"testing"
)

// TestIsNonInferencePath tests the non-inference path detection logic
func TestIsNonInferencePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// API endpoints (should be non-inference)
		{"API maas-api tokens", "/maas-api/v1/tokens", true},
		{"API maas-api tiers", "/maas-api/v1/tiers", true},
		{"API models listing", "/v1/models", true},
		{"API models with trailing slash", "/v1/models/", true},
		
		// OAuth endpoints (should be non-inference)
		{"OAuth callback", "/oauth/callback", true},
		{"OAuth route", "/oauth2", true},
		
		// Health and root (should be non-inference)
		{"Health check", "/health", true},
		{"Root path", "/", true},
		
		// Inference paths (should NOT be non-inference)
		{"LLM inference completions", "/llm/facebook-opt-125m-simulated/v1/completions", false},
		{"LLM inference chat", "/llm/facebook-opt-125m-simulated/v1/chat/completions", false},
		{"ACME inference", "/acme-inc-models/acme-inc-model-1/v1/chat/completions", false},
		{"Ethan inference", "/ethan-group-models/ethan-group-model-1/v1/completions", false},
		{"Serverless inference", "/serverless-models/my-model-name/v1/completions", false},
		
		// Unknown paths (should NOT be non-inference)
		{"Unknown path", "/unknown", false},
		{"Random path", "/some/random/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNonInferencePath(tt.path)
			if result != tt.expected {
				t.Errorf("isNonInferencePath(%s) = %v, want %v",
					tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsInferencePath tests the inference path detection logic
// Key pattern: /{namespace}/{model}/v1/{endpoint}
// The "v1" is the OpenAI API version that MaaS supports
func TestIsInferencePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Valid inference paths (all have /v1/ as 3rd segment)
		{"LLM namespace completions", "/llm/facebook-opt-125m-simulated/v1/completions", true},
		{"LLM namespace chat", "/llm/facebook-opt-125m-simulated/v1/chat/completions", true},
		{"ACME namespace", "/acme-inc-models/acme-inc-model-1/v1/chat/completions", true},
		{"Ethan namespace", "/ethan-group-models/ethan-group-model-1/v1/completions", true},
		{"Serverless namespace model 1", "/serverless-models/my-model-name/v1/completions", true},
		{"Serverless namespace model 2", "/serverless-models/serverless-model-1/v1/completions", true},
		{"Chat completions", "/llm/model/v1/chat/completions", true},
		{"Simple completions", "/namespace/model/v1/completions", true},
		
		// Non-inference paths (should be excluded)
		{"API endpoint", "/maas-api/v1/models", false},
		{"Models listing", "/v1/models", false},
		{"OAuth callback", "/oauth/callback", false},
		{"OAuth2 route", "/oauth2", false},
		{"Health check", "/health", false},
		{"Root", "/", false},
		
		// Invalid patterns (not enough segments or wrong structure)
		{"Two segments only", "/namespace/model", false},
		{"Wrong 3rd segment (v2)", "/namespace/model/v2/completions", false},
		{"Wrong 3rd segment (api)", "/namespace/model/api/completions", false},
		{"Single segment", "/single", false},
		{"Empty path", "", false},
		{"Just v1", "/v1", false},
		{"v1 with one segment", "/namespace/v1", false},
		
		// Edge cases
		{"Trailing slash", "/namespace/model/v1/completions/", true},
		{"Multiple slashes", "/namespace/model/v1/chat/completions/extra", true},
		{"v1 in wrong position", "/v1/namespace/model/completions", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInferencePath(tt.path)
			if result != tt.expected {
				t.Errorf("isInferencePath(%s) = %v, want %v",
					tt.path, result, tt.expected)
			}
		})
	}
}

// TestEvaluateRequest tests the request evaluation logic
func TestEvaluateRequest(t *testing.T) {
	tests := []struct {
		name          string
		allowValue    bool
		path          string
		method        string
		username      string
		expectedType  string
		expectedAllow bool
	}{
		// Non-inference requests (always allowed regardless of allowValue)
		{"API request allowed", true, "/maas-api/v1/models", "GET", "admin", "non-inference", true},
		{"API request allowed even when deny", false, "/maas-api/v1/models", "GET", "admin", "non-inference", true},
		{"Models listing allowed", true, "/v1/models", "GET", "user1", "non-inference", true},
		{"Models listing allowed even when deny", false, "/v1/models", "GET", "user1", "non-inference", true},
		{"OAuth allowed", true, "/oauth2", "GET", "user2", "non-inference", true},
		{"OAuth allowed even when deny", false, "/oauth2", "GET", "user2", "non-inference", true},
		
		// Inference requests with allowValue=true
		{"Inference LLM allowed", true, "/llm/facebook-opt-125m-simulated/v1/completions", "POST", "user1", "inference", true},
		{"Inference ACME allowed", true, "/acme-inc-models/acme-inc-model-1/v1/chat/completions", "POST", "user2", "inference", true},
		{"Inference serverless allowed", true, "/serverless-models/my-model/v1/completions", "POST", "user3", "inference", true},
		
		// Inference requests with allowValue=false (credit check denies)
		{"Inference LLM denied", false, "/llm/facebook-opt-125m-simulated/v1/completions", "POST", "user1", "inference", false},
		{"Inference ACME denied", false, "/acme-inc-models/acme-inc-model-1/v1/chat/completions", "POST", "user2", "inference", false},
		{"Inference serverless denied", false, "/serverless-models/my-model/v1/completions", "POST", "user3", "inference", false},
		
		// Unknown paths (always denied - fail-closed)
		{"Unknown path", true, "/unknown", "GET", "user1", "unknown", false},
		{"Random path", true, "/some/random/path", "POST", "user2", "unknown", false},
		{"Unknown even with allow", true, "/unknown/path", "GET", "user3", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CreditCheckHandler{allowValue: tt.allowValue}
			reqType, allow := handler.evaluateRequest(tt.path, tt.method, tt.username)
			
			if reqType != tt.expectedType {
				t.Errorf("evaluateRequest(%s) type = %s, want %s",
					tt.path, reqType, tt.expectedType)
			}
			if allow != tt.expectedAllow {
				t.Errorf("evaluateRequest(%s) allow = %v, want %v",
					tt.path, allow, tt.expectedAllow)
			}
		})
	}
}

// TestEvaluateRequestDifferentUsers tests that different users are handled correctly
func TestEvaluateRequestDifferentUsers(t *testing.T) {
	handler := &CreditCheckHandler{allowValue: true}
	
	users := []string{"admin", "user1", "alice", "bob", "system:serviceaccount:ns:sa"}
	
	for _, user := range users {
		t.Run("User: "+user, func(t *testing.T) {
			// Non-inference should always allow
			reqType, allow := handler.evaluateRequest("/v1/models", "GET", user)
			if reqType != "non-inference" || !allow {
				t.Errorf("Non-inference failed for user %s: type=%s allow=%v", user, reqType, allow)
			}
			
			// Inference should respect allowValue
			reqType, allow = handler.evaluateRequest("/llm/model/v1/completions", "POST", user)
			if reqType != "inference" || !allow {
				t.Errorf("Inference failed for user %s: type=%s allow=%v", user, reqType, allow)
			}
		})
	}
}

// TestEvaluateRequestAllowToggle tests toggling allowValue
func TestEvaluateRequestAllowToggle(t *testing.T) {
	path := "/llm/model/v1/completions"
	
	// Test with allowValue=true
	handlerAllow := &CreditCheckHandler{allowValue: true}
	reqType, allow := handlerAllow.evaluateRequest(path, "POST", "testuser")
	if reqType != "inference" || !allow {
		t.Errorf("With allowValue=true: got type=%s allow=%v, want type=inference allow=true",
			reqType, allow)
	}
	
	// Test with allowValue=false
	handlerDeny := &CreditCheckHandler{allowValue: false}
	reqType, allow = handlerDeny.evaluateRequest(path, "POST", "testuser")
	if reqType != "inference" || allow {
		t.Errorf("With allowValue=false: got type=%s allow=%v, want type=inference allow=false",
			reqType, allow)
	}
}

// TestInferencePathVariations tests various inference path variations
// to ensure the v1 detection is robust
func TestInferencePathVariations(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Should detect (v1 at position 2)
		{"/a/b/v1", true},
		{"/a/b/v1/", true},
		{"/a/b/v1/c", true},
		{"/a/b/v1/c/d", true},
		{"/a/b/v1/c/d/e", true},
		
		// Should NOT detect (v1 not at position 2)
		{"/a/v1", false},
		{"/a/v1/b", false},
		{"/a/b/c/v1", false},
		{"/v1/b/c", false},
		{"/a/b", false},
		{"/a/b/c", false},
		{"/a/b/v2/c", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isInferencePath(tt.path)
			if result != tt.expected {
				t.Errorf("isInferencePath(%s) = %v, want %v",
					tt.path, result, tt.expected)
			}
		})
	}
}

