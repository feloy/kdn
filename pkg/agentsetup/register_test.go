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

package agentsetup

import (
	"errors"
	"testing"

	"github.com/kortex-hub/kortex-cli/pkg/agent"
)

// fakeRegistrar implements AgentRegistrar for testing
type fakeRegistrar struct {
	registered map[string]agent.Agent
	failOn     string // agent name to fail registration on
}

func newFakeRegistrar() *fakeRegistrar {
	return &fakeRegistrar{
		registered: make(map[string]agent.Agent),
	}
}

func (f *fakeRegistrar) RegisterAgent(name string, ag agent.Agent) error {
	if name == f.failOn {
		return errors.New("registration failed")
	}
	f.registered[name] = ag
	return nil
}

func TestRegisterAll(t *testing.T) {
	t.Parallel()

	t.Run("registers all agents successfully", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()

		err := RegisterAll(registrar)
		if err != nil {
			t.Errorf("RegisterAll() error = %v, want nil", err)
		}

		// Verify claude agent was registered
		claudeAgent, exists := registrar.registered["claude"]
		if !exists {
			t.Fatal("claude agent was not registered")
		}

		if claudeAgent.Name() != "claude" {
			t.Errorf("claude agent name = %q, want %q", claudeAgent.Name(), "claude")
		}
	})

	t.Run("returns error if registration fails", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()
		registrar.failOn = "claude"

		err := RegisterAll(registrar)
		if err == nil {
			t.Error("RegisterAll() with failing registrar should return error")
		}
	})
}

func TestRegisterAllWithFactories(t *testing.T) {
	t.Parallel()

	t.Run("registers agents from custom factories", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()

		// Create custom factories
		factories := []agentFactory{
			agent.NewClaude,
		}

		err := registerAllWithFactories(registrar, factories)
		if err != nil {
			t.Errorf("registerAllWithFactories() error = %v, want nil", err)
		}

		if len(registrar.registered) != 1 {
			t.Errorf("registered %d agents, want 1", len(registrar.registered))
		}

		if _, exists := registrar.registered["claude"]; !exists {
			t.Error("claude agent was not registered")
		}
	})

	t.Run("handles empty factory list", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()
		factories := []agentFactory{}

		err := registerAllWithFactories(registrar, factories)
		if err != nil {
			t.Errorf("registerAllWithFactories() with empty list error = %v, want nil", err)
		}

		if len(registrar.registered) != 0 {
			t.Errorf("registered %d agents, want 0", len(registrar.registered))
		}
	})

	t.Run("stops on first registration error", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()
		registrar.failOn = "claude"

		factories := []agentFactory{
			agent.NewClaude,
		}

		err := registerAllWithFactories(registrar, factories)
		if err == nil {
			t.Error("registerAllWithFactories() should return error when registration fails")
		}
	})
}

func TestAvailableAgentsNotEmpty(t *testing.T) {
	t.Parallel()

	if len(availableAgents) == 0 {
		t.Error("availableAgents should not be empty")
	}

	// Verify all factories return valid agents
	for i, factory := range availableAgents {
		ag := factory()
		if ag == nil {
			t.Errorf("factory at index %d returned nil agent", i)
			continue
		}

		if ag.Name() == "" {
			t.Errorf("factory at index %d returned agent with empty name", i)
		}
	}
}
