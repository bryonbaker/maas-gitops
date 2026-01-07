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
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(allowValue string) *gin.Engine {
	// Set Gin to release mode to reduce logging noise
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Use custom logger middleware that doesn't duplicate our structured logs
	router.Use(gin.Recovery())

	// Determine allow boolean value
	allow := allowValue == "true"

	// Create handler
	creditHandler := NewCreditCheckHandler(allow)

	// Routes
	router.GET("/allow", creditHandler.HandleAllow)
	router.GET("/health", HandleHealth)

	return router
}

