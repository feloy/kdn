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
	"errors"
	"strings"
	"testing"

	"github.com/openkaiden/kdn/pkg/cmd/testutil"
	"github.com/openkaiden/kdn/pkg/provider"
	"github.com/openkaiden/kdn/pkg/providerservicesetup"
	"github.com/spf13/cobra"
)

// fakeProviderStore records calls for assertion in tests.
type fakeProviderStore struct {
	createParams provider.CreateParams
	createErr    error
	listItems    []provider.ListItem
	listErr      error
	removeName   string
	removeErr    error
}

func (f *fakeProviderStore) Create(params provider.CreateParams) error {
	f.createParams = params
	return f.createErr
}

func (f *fakeProviderStore) List() ([]provider.ListItem, error) {
	return f.listItems, f.listErr
}

func (f *fakeProviderStore) Get(name string) (provider.ListItem, map[string]string, error) {
	return provider.ListItem{}, nil, errors.New("not found")
}

func (f *fakeProviderStore) Remove(name string) error {
	f.removeName = name
	return f.removeErr
}

// buildProviderPreRunCmd creates a cobra.Command that mirrors the flag set seen by
// preRun when called through the real command tree.
func buildProviderPreRunCmd(storageDir string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("storage", storageDir, "")
	// Register all provider param flags so Changed() / GetString() work in tests
	registerProviderParamFlags(cmd, providerservicesetup.ListServices())
	return cmd
}

func TestBuildProviderCreateUsage(t *testing.T) {
	t.Parallel()

	usage := buildProviderCreateUsage(providerservicesetup.ListServices())

	// Must start with the command synopsis.
	if !strings.HasPrefix(usage, "create <name>") {
		t.Errorf("expected usage to start with 'create <name>', got: %s", usage)
	}
	// Types must appear in alphabetical order (anthropic before vertexai).
	anthropicIdx := strings.Index(usage, "--type anthropic")
	vertexaiIdx := strings.Index(usage, "--type vertexai")
	if anthropicIdx == -1 || vertexaiIdx == -1 {
		t.Fatalf("expected both --type anthropic and --type vertexai in usage, got: %s", usage)
	}
	if anthropicIdx > vertexaiIdx {
		t.Errorf("expected anthropic before vertexai in usage, got: %s", usage)
	}
	// Required anthropic param must appear without brackets.
	if !strings.Contains(usage, "--token <secret>") {
		t.Errorf("expected '--token <secret>' (required, no brackets) in usage, got: %s", usage)
	}
	// Optional anthropic param must appear with brackets.
	if !strings.Contains(usage, "[--url <url>]") {
		t.Errorf("expected '[--url <url>]' (optional, bracketed) in usage, got: %s", usage)
	}
	// Credential param must use <name-path> placeholder.
	if !strings.Contains(usage, "[--credentials <credentials-path>]") {
		t.Errorf("expected '[--credentials <credentials-path>]' in usage, got: %s", usage)
	}
	// Alternatives must be separated by |.
	if !strings.Contains(usage, "|") {
		t.Errorf("expected '|' separator between type alternatives, got: %s", usage)
	}
}

func TestProviderCreateCmd(t *testing.T) {
	t.Parallel()

	cmd := NewProviderCreateCmd()
	if cmd == nil {
		t.Fatal("NewProviderCreateCmd() returned nil")
	}
	if cmd.Use != "create <name>" {
		t.Errorf("expected Use %q, got %q", "create <name>", cmd.Use)
	}
}

func TestProviderCreateCmd_Examples(t *testing.T) {
	t.Parallel()

	cmd := NewProviderCreateCmd()
	if cmd.Example == "" {
		t.Fatal("Example field should not be empty")
	}

	commands, err := testutil.ParseExampleCommands(cmd.Example)
	if err != nil {
		t.Fatalf("failed to parse examples: %v", err)
	}

	expectedCount := 3
	if len(commands) != expectedCount {
		t.Errorf("expected %d example commands, got %d", expectedCount, len(commands))
	}

	rootCmd := NewRootCmd()
	if err := testutil.ValidateCommandExamples(rootCmd, cmd.Example); err != nil {
		t.Errorf("example validation failed: %v", err)
	}
}

func TestProviderCreateCmd_PreRun_InvalidOutput(t *testing.T) {
	t.Parallel()

	c := &providerCreateCmd{output: "xml", validTypes: []string{"anthropic", "vertexai"}}
	cmd := buildProviderPreRunCmd(t.TempDir())
	if err := c.preRun(cmd, []string{"name"}); err == nil {
		t.Fatal("expected error for unsupported output format")
	}
}

func TestProviderCreateCmd_PreRun(t *testing.T) {
	t.Parallel()

	t.Run("missing --type", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		err := c.preRun(cmd, []string{"name"})
		if err == nil || !strings.Contains(err.Error(), "--type is required") {
			t.Errorf("expected '--type is required' error, got: %v", err)
		}
	})

	t.Run("invalid --type", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "openai", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		err := c.preRun(cmd, []string{"name"})
		if err == nil || !strings.Contains(err.Error(), "invalid --type") {
			t.Errorf("expected 'invalid --type' error, got: %v", err)
		}
	})

	t.Run("anthropic missing --token", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "anthropic", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		err := c.preRun(cmd, []string{"name"})
		if err == nil || !strings.Contains(err.Error(), "--token is required") {
			t.Errorf("expected '--token is required' error, got: %v", err)
		}
	})

	t.Run("vertexai missing --project", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "vertexai", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		if err := cmd.Flags().Set("region", "us-central1"); err != nil {
			t.Fatal(err)
		}
		err := c.preRun(cmd, []string{"name"})
		if err == nil || !strings.Contains(err.Error(), "--project is required") {
			t.Errorf("expected '--project is required' error, got: %v", err)
		}
	})

	t.Run("vertexai missing --region", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "vertexai", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		if err := cmd.Flags().Set("project", "my-project"); err != nil {
			t.Fatal(err)
		}
		err := c.preRun(cmd, []string{"name"})
		if err == nil || !strings.Contains(err.Error(), "--region is required") {
			t.Errorf("expected '--region is required' error, got: %v", err)
		}
	})

	t.Run("anthropic with vertexai-only flag rejected", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "anthropic", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		if err := cmd.Flags().Set("token", "my-token"); err != nil {
			t.Fatal(err)
		}
		if err := cmd.Flags().Set("project", "my-project"); err != nil {
			t.Fatal(err)
		}
		err := c.preRun(cmd, []string{"name"})
		if err == nil || !strings.Contains(err.Error(), "--project is not valid for --type=anthropic") {
			t.Errorf("expected '--project is not valid' error, got: %v", err)
		}
	})

	t.Run("valid anthropic params", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "anthropic", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		if err := cmd.Flags().Set("token", "sk-ant-token"); err != nil {
			t.Fatal(err)
		}
		if err := c.preRun(cmd, []string{"my-provider"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if c.store == nil {
			t.Error("expected store to be initialised")
		}
		if len(c.params) != 1 || c.params[0].Name != "token" {
			t.Errorf("expected params to contain token, got: %v", c.params)
		}
	})

	t.Run("json output silences errors", func(t *testing.T) {
		t.Parallel()

		c := &providerCreateCmd{providerType: "anthropic", output: "json", validTypes: []string{"anthropic", "vertexai"}}
		cmd := buildProviderPreRunCmd(t.TempDir())
		if err := cmd.Flags().Set("token", "sk-ant-token"); err != nil {
			t.Fatal(err)
		}
		if err := c.preRun(cmd, []string{"my-provider"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !cmd.SilenceErrors {
			t.Error("expected cmd.SilenceErrors to be true when output is json")
		}
	})
}

func TestProviderCreateCmd_Run(t *testing.T) {
	t.Parallel()

	t.Run("calls store with correct params", func(t *testing.T) {
		t.Parallel()

		fs := &fakeProviderStore{}
		c := &providerCreateCmd{
			providerType: "anthropic",
			store:        fs,
			params: []provider.ProviderParamEntry{
				{Name: "token", Value: "sk-ant-token"},
			},
			validTypes: []string{"anthropic", "vertexai"},
		}

		root := &cobra.Command{}
		var out strings.Builder
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{"my-anthropic"}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}

		if fs.createParams.Name != "my-anthropic" {
			t.Errorf("Name: want %q, got %q", "my-anthropic", fs.createParams.Name)
		}
		if fs.createParams.Type != "anthropic" {
			t.Errorf("Type: want %q, got %q", "anthropic", fs.createParams.Type)
		}
		if len(fs.createParams.Params) != 1 || fs.createParams.Params[0].Name != "token" {
			t.Errorf("Params: want [{token}], got %v", fs.createParams.Params)
		}
		if !strings.Contains(out.String(), "my-anthropic") {
			t.Errorf("expected success message to contain provider name, got: %q", out.String())
		}
	})

	t.Run("store error propagates", func(t *testing.T) {
		t.Parallel()

		fs := &fakeProviderStore{createErr: errors.New("disk full")}
		c := &providerCreateCmd{store: fs, validTypes: []string{"anthropic", "vertexai"}}

		cmd := &cobra.Command{}
		err := c.run(cmd, []string{"x"})
		if err == nil {
			t.Fatal("expected error when store fails")
		}
	})

	t.Run("json output contains provider name", func(t *testing.T) {
		t.Parallel()

		fs := &fakeProviderStore{}
		c := &providerCreateCmd{
			providerType: "anthropic",
			output:       "json",
			store:        fs,
			validTypes:   []string{"anthropic", "vertexai"},
		}

		root := &cobra.Command{}
		var out bytes.Buffer
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{"my-anthropic"}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, `"name"`) {
			t.Errorf("expected 'name' key in JSON output, got: %s", output)
		}
		if !strings.Contains(output, `"my-anthropic"`) {
			t.Errorf("expected provider name in JSON output, got: %s", output)
		}
	})
}

func TestProviderCreateCmd_CredentialsFlagCompletion(t *testing.T) {
	t.Parallel()

	cmd := NewProviderCreateCmd()
	completionFunc, ok := cmd.GetFlagCompletionFunc("credentials")
	if !ok {
		t.Fatal("expected completion function registered for --credentials flag")
	}

	_, directive := completionFunc(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveDefault {
		t.Errorf("expected ShellCompDirectiveDefault for credential flag, got %v", directive)
	}
}

func TestProviderCreateCmd_TypeFlagCompletion(t *testing.T) {
	t.Parallel()

	cmd := NewProviderCreateCmd()
	completionFunc, ok := cmd.GetFlagCompletionFunc("type")
	if !ok {
		t.Fatal("expected completion function registered for --type flag")
	}

	completions, directive := completionFunc(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
	found := false
	for _, c := range completions {
		if c == "anthropic" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'anthropic' in completions, got %v", completions)
	}
}
