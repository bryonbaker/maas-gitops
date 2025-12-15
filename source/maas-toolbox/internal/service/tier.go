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

package service

import (
	"fmt"
	"maas-toolbox/internal/models"
	"maas-toolbox/internal/storage"
)

// TierService provides business logic for tier management
type TierService struct {
	storage *storage.K8sTierStorage
}

// NewTierService creates a new TierService instance
func NewTierService(storage *storage.K8sTierStorage) *TierService {
	return &TierService{
		storage: storage,
	}
}

// validateGroupsExist checks if all groups in the provided list exist in the cluster
func (s *TierService) validateGroupsExist(groups []string) error {
	for _, group := range groups {
		exists, err := s.storage.GroupExists(group)
		if err != nil {
			return fmt.Errorf("failed to check if group %s exists: %w", group, err)
		}
		if !exists {
			return models.ErrGroupNotFoundInCluster
		}
	}
	return nil
}

// CreateTier creates a new tier
func (s *TierService) CreateTier(tier *models.Tier) error {
	// Validate tier
	if err := tier.Validate(); err != nil {
		return err
	}

	// Validate all groups exist in cluster
	if len(tier.Groups) > 0 {
		if err := s.validateGroupsExist(tier.Groups); err != nil {
			return err
		}
	}

	// Load existing config
	config, err := s.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if tier already exists
	for _, existingTier := range config.Tiers {
		if existingTier.Name == tier.Name {
			return models.ErrTierAlreadyExists
		}
	}

	// Add new tier
	config.Tiers = append(config.Tiers, *tier)

	// Save config
	if err := s.storage.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetTiers returns all tiers
func (s *TierService) GetTiers() ([]models.Tier, error) {
	config, err := s.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return config.Tiers, nil
}

// GetTier returns a specific tier by name
func (s *TierService) GetTier(name string) (*models.Tier, error) {
	config, err := s.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, tier := range config.Tiers {
		if tier.Name == name {
			return &tier, nil
		}
	}

	return nil, models.ErrTierNotFound
}

// UpdateTier updates an existing tier
// Name cannot be changed, but description, level, and groups can be updated
func (s *TierService) UpdateTier(name string, updates *models.Tier) error {
	// Load existing config
	config, err := s.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find the tier
	var found bool
	for i := range config.Tiers {
		if config.Tiers[i].Name == name {
			// Ensure name is not being changed
			if updates.Name != "" && updates.Name != name {
				return models.ErrTierNameImmutable
			}

			// Update fields (only if provided)
			if updates.Description != "" {
				config.Tiers[i].Description = updates.Description
			}
			if updates.Level >= 0 {
				config.Tiers[i].Level = updates.Level
			}
			if updates.Groups != nil {
				// Validate all groups before updating
				for _, group := range updates.Groups {
					if err := models.ValidateGroupName(group); err != nil {
						return err
					}
				}
				// Validate all groups exist in cluster
				if err := s.validateGroupsExist(updates.Groups); err != nil {
					return err
				}
				config.Tiers[i].Groups = updates.Groups
			}

			// Validate updated tier
			if err := config.Tiers[i].Validate(); err != nil {
				return err
			}

			found = true
			break
		}
	}

	if !found {
		return models.ErrTierNotFound
	}

	// Save config
	if err := s.storage.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// DeleteTier deletes a tier by name
func (s *TierService) DeleteTier(name string) error {
	// Load existing config
	config, err := s.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find and remove the tier
	var found bool
	for i, tier := range config.Tiers {
		if tier.Name == name {
			config.Tiers = append(config.Tiers[:i], config.Tiers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return models.ErrTierNotFound
	}

	// Save config
	if err := s.storage.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// AddGroup adds a group to a tier
func (s *TierService) AddGroup(tierName, groupName string) error {
	// Validate group name format
	if err := models.ValidateGroupName(groupName); err != nil {
		return err
	}

	// Validate group exists in cluster
	exists, err := s.storage.GroupExists(groupName)
	if err != nil {
		return fmt.Errorf("failed to check if group exists: %w", err)
	}
	if !exists {
		return models.ErrGroupNotFoundInCluster
	}

	// Load existing config
	config, err := s.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find the tier
	var found bool
	for i := range config.Tiers {
		if config.Tiers[i].Name == tierName {
			// Check if group already exists
			for _, existingGroup := range config.Tiers[i].Groups {
				if existingGroup == groupName {
					return models.ErrGroupAlreadyExists
				}
			}

			// Add the group
			config.Tiers[i].Groups = append(config.Tiers[i].Groups, groupName)
			found = true
			break
		}
	}

	if !found {
		return models.ErrTierNotFound
	}

	// Save config
	if err := s.storage.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RemoveGroup removes a group from a tier
func (s *TierService) RemoveGroup(tierName, groupName string) error {
	// Validate group name format
	if err := models.ValidateGroupName(groupName); err != nil {
		return err
	}

	// Load existing config
	config, err := s.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find the tier
	var found bool
	var groupFound bool
	for i := range config.Tiers {
		if config.Tiers[i].Name == tierName {
			found = true
			// Find and remove the group
			for j, group := range config.Tiers[i].Groups {
				if group == groupName {
					config.Tiers[i].Groups = append(config.Tiers[i].Groups[:j], config.Tiers[i].Groups[j+1:]...)
					groupFound = true
					break
				}
			}
			break
		}
	}

	if !found {
		return models.ErrTierNotFound
	}

	if !groupFound {
		return models.ErrGroupNotFound
	}

	// Save config
	if err := s.storage.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetTiersByGroup returns all tiers that contain the specified group
func (s *TierService) GetTiersByGroup(groupName string) ([]models.Tier, error) {
	// Validate group name format
	if err := models.ValidateGroupName(groupName); err != nil {
		return nil, err
	}

	// Load all tiers
	config, err := s.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Filter tiers that contain the specified group
	var matchingTiers []models.Tier
	for _, tier := range config.Tiers {
		for _, group := range tier.Groups {
			if group == groupName {
				matchingTiers = append(matchingTiers, tier)
				break
			}
		}
	}

	return matchingTiers, nil
}
