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
	"strings"
	"testing"
)

func TestProviderCmd(t *testing.T) {
	t.Parallel()

	cmd := NewProviderCmd()
	if cmd == nil {
		t.Fatal("NewProviderCmd() returned nil")
	}
	if cmd.Use != "provider" {
		t.Errorf("expected Use %q, got %q", "provider", cmd.Use)
	}

	subCmds := cmd.Commands()
	if len(subCmds) == 0 {
		t.Fatal("expected provider command to have subcommands")
	}

	foundCreate := false
	foundList := false
	foundRemove := false
	for _, sub := range subCmds {
		switch sub.Use {
		case "create <name>":
			foundCreate = true
		case "list":
			foundList = true
		case "remove <name>":
			foundRemove = true
		}
	}
	if !foundCreate {
		t.Error("expected provider command to have 'create' subcommand")
	}
	if !foundList {
		t.Error("expected provider command to have 'list' subcommand")
	}
	if !foundRemove {
		t.Error("expected provider command to have 'remove' subcommand")
	}
}

func TestProviderCmd_UnknownCommand(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"provider", "foobar"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' in error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "foobar") {
		t.Errorf("expected 'foobar' in error, got: %s", err.Error())
	}
}
