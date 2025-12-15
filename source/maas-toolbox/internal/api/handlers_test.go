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
	"bytes"
	"encoding/json"
	"maas-toolbox/internal/models"
	"maas-toolbox/internal/service"
	"maas-toolbox/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes/fake"
)

// createEmptyMockK8sStorage creates a mock storage with no ConfigMap (will return empty)
func createEmptyMockK8sStorage() *storage.K8sTierStorage {
	client := fake.NewSimpleClientset()
	return storage.NewK8sTierStorage(client, "test", "tier-to-group-mapping")
}

func setupTestRouter() (*gin.Engine, *TierHandler) {
	gin.SetMode(gin.TestMode)
	mockStore := createEmptyMockK8sStorage()
	tierService := service.NewTierService(mockStore)
	handler := NewTierHandler(tierService)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/tiers", handler.CreateTier)
		v1.GET("/tiers", handler.GetTiers)
		v1.GET("/tiers/:name", handler.GetTier)
		v1.PUT("/tiers/:name", handler.UpdateTier)
		v1.DELETE("/tiers/:name", handler.DeleteTier)
		v1.POST("/tiers/:name/groups", handler.AddGroup)
		v1.DELETE("/tiers/:name/groups/:group", handler.RemoveGroup)
		v1.GET("/groups/:group/tiers", handler.GetTiersByGroup)
	}
	return router, handler
}

func TestCreateTier_WithoutGroups(t *testing.T) {
	router, _ := setupTestRouter()

	// Test creating a tier without groups field (should default to empty list)
	tierJSON := `{
		"name": "test-tier",
		"description": "Test tier without groups",
		"level": 1
	}`

	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tierJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.Tier
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "test-tier" {
		t.Errorf("Expected name 'test-tier', got '%s'", response.Name)
	}
	if response.Description != "Test tier without groups" {
		t.Errorf("Expected description 'Test tier without groups', got '%s'", response.Description)
	}
	if response.Level != 1 {
		t.Errorf("Expected level 1, got %d", response.Level)
	}
	if response.Groups == nil {
		t.Error("Groups should not be nil")
	}
	if len(response.Groups) != 0 {
		t.Errorf("Groups should be empty, got length %d", len(response.Groups))
	}
	if response.Groups == nil || len(response.Groups) != 0 {
		t.Errorf("Groups should be an empty list [], got %v", response.Groups)
	}
}

func TestCreateTier_WithEmptyGroups(t *testing.T) {
	router, _ := setupTestRouter()

	// Test creating a tier with explicit empty groups array
	tierJSON := `{
		"name": "test-tier-empty",
		"description": "Test tier with empty groups",
		"level": 1,
		"groups": []
	}`

	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tierJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.Tier
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "test-tier-empty" {
		t.Errorf("Expected name 'test-tier-empty', got '%s'", response.Name)
	}
	if response.Groups == nil {
		t.Error("Groups should not be nil")
	}
	if len(response.Groups) != 0 {
		t.Errorf("Groups should be empty, got length %d", len(response.Groups))
	}
}

func TestCreateTier_WithGroups(t *testing.T) {
	router, _ := setupTestRouter()

	// Test creating a tier with groups provided
	tierJSON := `{
		"name": "test-tier-with-groups",
		"description": "Test tier with groups",
		"level": 1,
		"groups": ["system:authenticated", "premium-users"]
	}`

	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tierJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.Tier
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "test-tier-with-groups" {
		t.Errorf("Expected name 'test-tier-with-groups', got '%s'", response.Name)
	}
	if response.Groups == nil {
		t.Error("Groups should not be nil")
	}
	expectedGroups := []string{"system:authenticated", "premium-users"}
	if len(response.Groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(response.Groups))
	}
	for i, group := range expectedGroups {
		if i >= len(response.Groups) || response.Groups[i] != group {
			t.Errorf("Expected group[%d] to be '%s', got '%v'", i, group, response.Groups)
		}
	}
}

func TestCreateTier_VerifyGroupsDefaultedInStorage(t *testing.T) {
	mockStore := createEmptyMockK8sStorage()
	tierService := service.NewTierService(mockStore)
	handler := NewTierHandler(tierService)
	router := gin.New()
	gin.SetMode(gin.TestMode)
	v1 := router.Group("/api/v1")
	v1.POST("/tiers", handler.CreateTier)

	// Create tier without groups
	tierJSON := `{
		"name": "stored-tier",
		"description": "Tier to verify storage",
		"level": 2
	}`

	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tierJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify the tier was stored with empty groups
	config, err := mockStore.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if len(config.Tiers) != 1 {
		t.Errorf("Expected 1 tier in storage, got %d", len(config.Tiers))
	}
	if config.Tiers[0].Name != "stored-tier" {
		t.Errorf("Expected tier name 'stored-tier', got '%s'", config.Tiers[0].Name)
	}
	if config.Tiers[0].Groups == nil {
		t.Error("Groups should not be nil in storage")
	}
	if len(config.Tiers[0].Groups) != 0 {
		t.Errorf("Groups should be empty list in storage, got length %d", len(config.Tiers[0].Groups))
	}
}

func TestGetTiersByGroup_SingleTier(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a tier with a group
	tierJSON := `{
		"name": "premium",
		"description": "Premium tier",
		"level": 3,
		"groups": ["premium-users"]
	}`

	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tierJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create tier: expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Get tiers by group
	req, _ = http.NewRequest("GET", "/api/v1/groups/premium-users/tiers", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []models.Tier
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Expected 1 tier, got %d", len(response))
	}
	if response[0].Name != "premium" {
		t.Errorf("Expected tier name 'premium', got '%s'", response[0].Name)
	}
}

func TestGetTiersByGroup_MultipleTiers(t *testing.T) {
	router, _ := setupTestRouter()

	// Create multiple tiers with the same group
	tier1JSON := `{
		"name": "premium",
		"description": "Premium tier",
		"level": 3,
		"groups": ["premium-users", "vip-users"]
	}`
	tier2JSON := `{
		"name": "enterprise",
		"description": "Enterprise tier",
		"level": 4,
		"groups": ["premium-users", "enterprise-users"]
	}`

	// Create first tier
	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tier1JSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create first tier: expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Create second tier
	req, _ = http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tier2JSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create second tier: expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Get tiers by group
	req, _ = http.NewRequest("GET", "/api/v1/groups/premium-users/tiers", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []models.Tier
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 tiers, got %d", len(response))
	}

	// Verify both tiers are in the response
	tierNames := make(map[string]bool)
	for _, tier := range response {
		tierNames[tier.Name] = true
	}
	if !tierNames["premium"] {
		t.Error("Expected 'premium' tier in response")
	}
	if !tierNames["enterprise"] {
		t.Error("Expected 'enterprise' tier in response")
	}
}

func TestGetTiersByGroup_NoMatchingTiers(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a tier with different groups
	tierJSON := `{
		"name": "free",
		"description": "Free tier",
		"level": 1,
		"groups": ["system:authenticated"]
	}`

	req, _ := http.NewRequest("POST", "/api/v1/tiers", bytes.NewBufferString(tierJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create tier: expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Get tiers by group that doesn't exist in any tier
	req, _ = http.NewRequest("GET", "/api/v1/groups/nonexistent-group/tiers", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []models.Tier
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected 0 tiers, got %d", len(response))
	}
}

func TestGetTiersByGroup_InvalidGroupName(t *testing.T) {
	router, _ := setupTestRouter()

	// Try to get tiers with invalid group name (contains uppercase)
	req, _ := http.NewRequest("GET", "/api/v1/groups/InvalidGroup/tiers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error == "" {
		t.Error("Expected error message in response")
	}
}
