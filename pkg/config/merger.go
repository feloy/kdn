// Copyright 2026 Red Hat, Inc.
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

package config

import (
	workspace "github.com/openkaiden/kdn-api/workspace-configuration/go"
)

// Merger merges multiple WorkspaceConfiguration objects with proper precedence rules.
// When merging:
// - Environment variables: Later configs override earlier ones (by name)
// - Mounts: Deduplicated by host+target pair (preserves order, no duplicates)
type Merger interface {
	// Merge combines two WorkspaceConfiguration objects.
	// The override config takes precedence over the base config.
	// Returns a new merged configuration without modifying the inputs.
	Merge(base, override *workspace.WorkspaceConfiguration) *workspace.WorkspaceConfiguration
}

// merger is the internal implementation of Merger
type merger struct{}

// Compile-time check to ensure merger implements Merger interface
var _ Merger = (*merger)(nil)

// NewMerger creates a new configuration merger
func NewMerger() Merger {
	return &merger{}
}

// Merge combines two WorkspaceConfiguration objects with override taking precedence
func (m *merger) Merge(base, override *workspace.WorkspaceConfiguration) *workspace.WorkspaceConfiguration {
	// If both are nil, return nil
	if base == nil && override == nil {
		return nil
	}

	// If only base is nil, return a copy of override
	if base == nil {
		return copyConfig(override)
	}

	// If only override is nil, return a copy of base
	if override == nil {
		return copyConfig(base)
	}

	// Merge both configurations
	result := &workspace.WorkspaceConfiguration{}

	// Merge environment variables
	result.Environment = mergeEnvironment(base.Environment, override.Environment)

	// Merge mounts
	result.Mounts = mergeMounts(base.Mounts, override.Mounts)

	// Merge skills
	result.Skills = mergeSkills(base.Skills, override.Skills)

	// Merge MCP configuration
	result.Mcp = mergeMCP(base.Mcp, override.Mcp)

	return result
}

// mergeEnvironment merges environment variables, with override taking precedence by name
func mergeEnvironment(base, override *[]workspace.EnvironmentVariable) *[]workspace.EnvironmentVariable {
	if base == nil && override == nil {
		return nil
	}

	// Create a map to track variables by name
	envMap := make(map[string]workspace.EnvironmentVariable)
	var order []string

	// Add base environment variables
	if base != nil {
		for _, env := range *base {
			envMap[env.Name] = env
			order = append(order, env.Name)
		}
	}

	// Override with variables from override config
	if override != nil {
		for _, env := range *override {
			if _, exists := envMap[env.Name]; !exists {
				// New variable, add to order
				order = append(order, env.Name)
			}
			// Override or add the variable
			envMap[env.Name] = env
		}
	}

	// Build result array preserving order
	if len(envMap) == 0 {
		return nil
	}

	result := make([]workspace.EnvironmentVariable, 0, len(order))
	for _, name := range order {
		result = append(result, envMap[name])
	}

	return &result
}

// deepCopyMount returns a deep copy of m with the Ro pointer independent from the original.
func deepCopyMount(m workspace.Mount) workspace.Mount {
	if m.Ro != nil {
		roCopy := *m.Ro
		m.Ro = &roCopy
	}
	return m
}

// mergeMounts merges mount slices, deduplicating by host+target pair.
// Mounts from base are appended first; if override contains a mount with the same
// host+target key, it replaces the base entry in-place (preserving position) so that
// per-mount fields such as Ro are correctly overridden.
func mergeMounts(base, override *[]workspace.Mount) *[]workspace.Mount {
	if base == nil && override == nil {
		return nil
	}

	type mountKey struct{ host, target string }
	seen := make(map[mountKey]int) // value is index in result
	var result []workspace.Mount

	for _, slice := range []*[]workspace.Mount{base, override} {
		if slice == nil {
			continue
		}
		isOverride := slice == override
		for _, m := range *slice {
			key := mountKey{m.Host, m.Target}
			if idx, exists := seen[key]; !exists {
				seen[key] = len(result)
				result = append(result, deepCopyMount(m))
			} else if isOverride {
				result[idx] = deepCopyMount(m)
			}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return &result
}

// mergeSkills merges skills slices, deduplicating by path value.
// Skills from base come first; skills from override are appended if not already present.
func mergeSkills(base, override *[]string) *[]string {
	if base == nil && override == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []string

	for _, slice := range []*[]string{base, override} {
		if slice == nil {
			continue
		}
		for _, s := range *slice {
			if !seen[s] {
				seen[s] = true
				result = append(result, s)
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return &result
}

// mergeMCP merges two McpConfiguration objects, with override taking precedence by name.
// Commands and servers from base are included first; override entries replace base entries
// with the same name.
func mergeMCP(base, override *workspace.McpConfiguration) *workspace.McpConfiguration {
	if base == nil && override == nil {
		return nil
	}
	if base == nil {
		return copyMCP(override)
	}
	if override == nil {
		return copyMCP(base)
	}

	result := &workspace.McpConfiguration{}
	result.Commands = mergeMCPCommands(base.Commands, override.Commands)
	result.Servers = mergeMCPServers(base.Servers, override.Servers)

	if result.Commands == nil && result.Servers == nil {
		return nil
	}
	return result
}

// mergeMCPCommands merges command slices, deduplicating by name (override wins).
func mergeMCPCommands(base, override *[]workspace.McpCommand) *[]workspace.McpCommand {
	if base == nil && override == nil {
		return nil
	}

	cmdMap := make(map[string]workspace.McpCommand)
	var order []string

	if base != nil {
		for _, cmd := range *base {
			cmdMap[cmd.Name] = cmd
			order = append(order, cmd.Name)
		}
	}
	if override != nil {
		for _, cmd := range *override {
			if _, exists := cmdMap[cmd.Name]; !exists {
				order = append(order, cmd.Name)
			}
			cmdMap[cmd.Name] = cmd
		}
	}

	if len(cmdMap) == 0 {
		return nil
	}

	result := make([]workspace.McpCommand, 0, len(order))
	for _, name := range order {
		result = append(result, cmdMap[name])
	}
	return &result
}

// mergeMCPServers merges server slices, deduplicating by name (override wins).
func mergeMCPServers(base, override *[]workspace.McpServer) *[]workspace.McpServer {
	if base == nil && override == nil {
		return nil
	}

	srvMap := make(map[string]workspace.McpServer)
	var order []string

	if base != nil {
		for _, srv := range *base {
			srvMap[srv.Name] = srv
			order = append(order, srv.Name)
		}
	}
	if override != nil {
		for _, srv := range *override {
			if _, exists := srvMap[srv.Name]; !exists {
				order = append(order, srv.Name)
			}
			srvMap[srv.Name] = srv
		}
	}

	if len(srvMap) == 0 {
		return nil
	}

	result := make([]workspace.McpServer, 0, len(order))
	for _, name := range order {
		result = append(result, srvMap[name])
	}
	return &result
}

// copyMCP creates a deep copy of an McpConfiguration.
func copyMCP(mcp *workspace.McpConfiguration) *workspace.McpConfiguration {
	if mcp == nil {
		return nil
	}
	result := &workspace.McpConfiguration{}
	if mcp.Commands != nil {
		cmdsCopy := make([]workspace.McpCommand, len(*mcp.Commands))
		copy(cmdsCopy, *mcp.Commands)
		result.Commands = &cmdsCopy
	}
	if mcp.Servers != nil {
		srvsCopy := make([]workspace.McpServer, len(*mcp.Servers))
		copy(srvsCopy, *mcp.Servers)
		result.Servers = &srvsCopy
	}
	return result
}

// copyConfig creates a deep copy of a WorkspaceConfiguration
func copyConfig(cfg *workspace.WorkspaceConfiguration) *workspace.WorkspaceConfiguration {
	if cfg == nil {
		return nil
	}

	result := &workspace.WorkspaceConfiguration{}

	// Copy environment variables
	if cfg.Environment != nil {
		envCopy := make([]workspace.EnvironmentVariable, len(*cfg.Environment))
		copy(envCopy, *cfg.Environment)
		result.Environment = &envCopy
	}

	// Copy mounts (deep copy each entry so Ro pointers are independent)
	if cfg.Mounts != nil {
		mountsCopy := make([]workspace.Mount, len(*cfg.Mounts))
		for i, m := range *cfg.Mounts {
			mountsCopy[i] = deepCopyMount(m)
		}
		result.Mounts = &mountsCopy
	}

	// Copy skills
	if cfg.Skills != nil {
		skillsCopy := make([]string, len(*cfg.Skills))
		copy(skillsCopy, *cfg.Skills)
		result.Skills = &skillsCopy
	}

	// Copy MCP configuration
	result.Mcp = copyMCP(cfg.Mcp)

	return result
}
