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

package models

import "errors"

var (
	ErrTierNameRequired        = errors.New("tier name is required")
	ErrTierDescriptionRequired = errors.New("tier description is required")
	ErrTierLevelInvalid        = errors.New("tier level must be non-negative")
	ErrTierNotFound            = errors.New("tier not found")
	ErrTierAlreadyExists       = errors.New("tier already exists")
	ErrTierNameImmutable       = errors.New("tier name cannot be changed")
	ErrGroupRequired           = errors.New("group name is required")
	ErrGroupAlreadyExists      = errors.New("group already exists in tier")
	ErrGroupNotFound           = errors.New("group not found in tier")
	ErrGroupNotFoundInCluster  = errors.New("group not found in cluster")
	ErrInvalidKubernetesName   = errors.New("invalid Kubernetes name format: must be 1-253 characters, start and end with alphanumeric, and contain only lowercase alphanumeric, hyphens, colons, dots, or underscores")
)

