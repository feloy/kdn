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

package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd_HasVersionCommand(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCmd()
	versionCmd := rootCmd.Commands()
	found := false
	for _, cmd := range versionCmd {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected rootCmd to have 'version' subcommand")
	}
}

func TestExecute_WithVersion(t *testing.T) {
	t.Parallel()

	// Redirect output to avoid cluttering test output
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	// Call Execute() and verify it succeeds
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
}
