/**********************************************************************
 * Copyright (C) 2026 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 **********************************************************************/

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	workspace "github.com/openkaiden/kdn-api/workspace-configuration/go"
)

// workspaceConfigFile is the on-disk representation of workspace.json. Embedding
// workspace.WorkspaceConfiguration promotes all its fields so callers can access
// them directly, while Schema carries the optional JSON Schema URL.
type workspaceConfigFile struct {
	Schema string `json:"$schema,omitempty"`
	workspace.WorkspaceConfiguration
}

// WorkspaceConfigUpdater manages the local workspace configuration file.
type WorkspaceConfigUpdater interface {
	// AddSecret appends secretName to the Secrets list of the workspace config,
	// creating the file and directory if they do not yet exist.
	// The call is idempotent: if the secret is already present it is not duplicated.
	AddSecret(secretName string) error

	// AddEnvVar adds or updates an environment variable in the workspace config.
	// If an entry with the same name already exists its value is replaced.
	AddEnvVar(name, value string) error

	// AddMount adds a mount entry to the workspace config.
	// The call is idempotent: if a mount with the same host and target already exists
	// it is not duplicated.
	AddMount(host, target string, ro bool) error

	// AddPort appends port to the Ports list of the workspace config.
	// The call is idempotent: if the port is already present it is not duplicated.
	AddPort(port int) error

	// AddFeature adds a feature entry to the Features map of the workspace config.
	// The call is idempotent: if the feature ID is already present it is not modified.
	AddFeature(featureID string, options map[string]interface{}) error
}

type workspaceConfigUpdater struct {
	configDir string // absolute path to the .kaiden/ directory
}

var _ WorkspaceConfigUpdater = (*workspaceConfigUpdater)(nil)

// NewWorkspaceConfigUpdater returns a WorkspaceConfigUpdater backed by
// <configDir>/workspace.json.
func NewWorkspaceConfigUpdater(configDir string) (WorkspaceConfigUpdater, error) {
	if configDir == "" {
		return nil, ErrInvalidPath
	}
	absPath, err := filepath.Abs(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config directory path: %w", err)
	}
	return &workspaceConfigUpdater{configDir: absPath}, nil
}

func (w *workspaceConfigUpdater) AddEnvVar(name, value string) error {
	configPath := filepath.Join(w.configDir, WorkspaceConfigFile)

	cfg, isNew, err := w.readConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.Environment == nil {
		v := value
		envVars := []workspace.EnvironmentVariable{{Name: name, Value: &v}}
		cfg.Environment = &envVars
	} else {
		for i, e := range *cfg.Environment {
			if e.Name == name {
				v := value
				(*cfg.Environment)[i].Value = &v
				(*cfg.Environment)[i].Secret = nil
				return w.writeConfig(configPath, cfg, isNew)
			}
		}
		v := value
		*cfg.Environment = append(*cfg.Environment, workspace.EnvironmentVariable{Name: name, Value: &v})
	}

	return w.writeConfig(configPath, cfg, isNew)
}

func (w *workspaceConfigUpdater) AddMount(host, target string, ro bool) error {
	configPath := filepath.Join(w.configDir, WorkspaceConfigFile)

	cfg, isNew, err := w.readConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.Mounts == nil {
		roVal := ro
		mounts := []workspace.Mount{{Host: host, Target: target, Ro: &roVal}}
		cfg.Mounts = &mounts
	} else {
		for _, m := range *cfg.Mounts {
			if m.Host == host && m.Target == target {
				return nil
			}
		}
		roVal := ro
		*cfg.Mounts = append(*cfg.Mounts, workspace.Mount{Host: host, Target: target, Ro: &roVal})
	}

	return w.writeConfig(configPath, cfg, isNew)
}

func (w *workspaceConfigUpdater) AddSecret(secretName string) error {
	configPath := filepath.Join(w.configDir, WorkspaceConfigFile)

	cfg, isNew, err := w.readConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.Secrets == nil {
		secrets := []string{secretName}
		cfg.Secrets = &secrets
	} else {
		for _, s := range *cfg.Secrets {
			if s == secretName {
				return nil
			}
		}
		*cfg.Secrets = append(*cfg.Secrets, secretName)
	}

	return w.writeConfig(configPath, cfg, isNew)
}

func (w *workspaceConfigUpdater) readConfig(configPath string) (*workspaceConfigFile, bool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &workspaceConfigFile{}, true, nil
		}
		return nil, false, fmt.Errorf("failed to read workspace config: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &workspaceConfigFile{}, false, nil
	}

	var cfg workspaceConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, false, fmt.Errorf("failed to parse workspace config: %w", err)
	}
	return &cfg, false, nil
}

func (w *workspaceConfigUpdater) AddPort(port int) error {
	configPath := filepath.Join(w.configDir, WorkspaceConfigFile)

	cfg, isNew, err := w.readConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.Ports == nil {
		ports := []int{port}
		cfg.Ports = &ports
	} else {
		for _, p := range *cfg.Ports {
			if p == port {
				return nil
			}
		}
		*cfg.Ports = append(*cfg.Ports, port)
	}

	return w.writeConfig(configPath, cfg, isNew)
}

func (w *workspaceConfigUpdater) AddFeature(featureID string, options map[string]interface{}) error {
	configPath := filepath.Join(w.configDir, WorkspaceConfigFile)

	cfg, isNew, err := w.readConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.Features == nil {
		features := map[string]map[string]interface{}{featureID: options}
		cfg.Features = &features
	} else {
		if _, exists := (*cfg.Features)[featureID]; exists {
			return nil
		}
		(*cfg.Features)[featureID] = options
	}

	return w.writeConfig(configPath, cfg, isNew)
}

func (w *workspaceConfigUpdater) writeConfig(configPath string, cfg *workspaceConfigFile, isNew bool) error {
	if isNew {
		cfg.Schema = WorkspaceSchemaURL
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write workspace config: %w", err)
	}
	return nil
}
