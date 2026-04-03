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

// Agent is an interface for agent-specific configuration and setup operations.
type Agent interface {
	// Name returns the agent name (e.g., "claude", "goose").
	Name() string
	// SkipOnboarding modifies agent settings to skip onboarding prompts.
	// It takes the current agent settings map (path -> content) and the workspace
	// sources path inside the container, and returns the modified settings with
	// onboarding flags set appropriately.
	// Returns the modified settings map, or an error if modification fails.
	SkipOnboarding(settings map[string][]byte, workspaceSourcesPath string) (map[string][]byte, error)
}
