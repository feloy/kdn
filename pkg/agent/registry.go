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

package agent

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrAgentNotFound is returned when an agent is not found in the registry.
	ErrAgentNotFound = errors.New("agent not found")
)

// Registry manages agent implementations.
type Registry interface {
	// Register registers an agent implementation by name.
	Register(name string, agent Agent) error
	// Get retrieves an agent implementation by name.
	// Returns ErrAgentNotFound if the agent is not registered.
	Get(name string) (Agent, error)
	// List returns all registered agent names.
	List() []string
}

// registry is the internal implementation of Registry.
type registry struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

// Compile-time check to ensure registry implements Registry interface
var _ Registry = (*registry)(nil)

// NewRegistry creates a new agent registry.
func NewRegistry() Registry {
	return &registry{
		agents: make(map[string]Agent),
	}
}

// Register registers an agent implementation by name.
func (r *registry) Register(name string, agent Agent) error {
	if name == "" {
		return errors.New("agent name cannot be empty")
	}
	if agent == nil {
		return errors.New("agent cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; exists {
		return fmt.Errorf("agent %q is already registered", name)
	}

	r.agents[name] = agent
	return nil
}

// Get retrieves an agent implementation by name.
func (r *registry) Get(name string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[name]
	if !exists {
		return nil, ErrAgentNotFound
	}

	return agent, nil
}

// List returns all registered agent names.
func (r *registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}
