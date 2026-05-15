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
	"github.com/openkaiden/kdn/pkg/providerservice"
	"github.com/spf13/cobra"
)

func TestProviderListCmd(t *testing.T) {
	t.Parallel()

	cmd := NewProviderListCmd()
	if cmd == nil {
		t.Fatal("NewProviderListCmd() returned nil")
	}
	if cmd.Use != "list" {
		t.Errorf("expected Use %q, got %q", "list", cmd.Use)
	}
}

func TestProviderListCmd_Examples(t *testing.T) {
	t.Parallel()

	cmd := NewProviderListCmd()
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

func TestProviderListCmd_PreRun(t *testing.T) {
	t.Parallel()

	c := &providerListCmd{}
	cmd := &cobra.Command{}
	cmd.Flags().String("storage", t.TempDir(), "")

	if err := c.preRun(cmd, []string{}); err != nil {
		t.Fatalf("preRun() failed: %v", err)
	}
	if c.store == nil {
		t.Error("expected store to be initialised")
	}
}

func TestProviderListCmd_PreRun_InvalidOutput(t *testing.T) {
	t.Parallel()

	c := &providerListCmd{output: "xml"}
	cmd := &cobra.Command{}
	cmd.Flags().String("storage", t.TempDir(), "")

	if err := c.preRun(cmd, []string{}); err == nil {
		t.Fatal("expected error for unsupported output format")
	}
}

func TestProviderListCmd_Run(t *testing.T) {
	t.Parallel()

	t.Run("displays empty message when no providers", func(t *testing.T) {
		t.Parallel()

		c := &providerListCmd{store: &fakeProviderStore{}}
		root := &cobra.Command{}
		var out bytes.Buffer
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}
		if !strings.Contains(out.String(), "No providers found") {
			t.Errorf("expected 'No providers found' in output, got: %s", out.String())
		}
	})

	t.Run("table output contains provider fields", func(t *testing.T) {
		t.Parallel()

		c := &providerListCmd{store: &fakeProviderStore{
			listItems: []provider.ListItem{
				{
					Name: "my-anthropic",
					Type: "anthropic",
					Params: []provider.ProviderParamEntry{
						{Name: "url", Kind: providerservice.ProviderParamKindURL, Value: "https://api.anthropic.com"},
					},
				},
			},
		}}
		root := &cobra.Command{}
		var out bytes.Buffer
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "my-anthropic") {
			t.Errorf("expected 'my-anthropic' in output, got: %s", output)
		}
		if !strings.Contains(output, "anthropic") {
			t.Errorf("expected 'anthropic' in output, got: %s", output)
		}
		if !strings.Contains(output, "url=https://api.anthropic.com") {
			t.Errorf("expected param in output, got: %s", output)
		}
	})

	t.Run("table output shows secret params as <secret>", func(t *testing.T) {
		t.Parallel()

		c := &providerListCmd{store: &fakeProviderStore{
			listItems: []provider.ListItem{
				{
					Name: "my-anthropic",
					Type: "anthropic",
					Params: []provider.ProviderParamEntry{
						// Value is the reference name returned by the real store.
						{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "my-anthropic/token"},
					},
				},
			},
		}}
		root := &cobra.Command{}
		var out bytes.Buffer
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "token=<secret>") {
			t.Errorf("expected 'token=<secret>' in table output, got: %s", output)
		}
		// The actual reference name must not be shown in the table.
		if strings.Contains(output, "my-anthropic/token") {
			t.Errorf("expected reference name to be hidden in table, got: %s", output)
		}
	})

	t.Run("json output contains all fields including secret reference name", func(t *testing.T) {
		t.Parallel()

		c := &providerListCmd{
			output: "json",
			store: &fakeProviderStore{
				listItems: []provider.ListItem{
					{
						Name: "my-anthropic",
						Type: "anthropic",
						Params: []provider.ProviderParamEntry{
							// Value is the reference name returned by the real store.
							{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "my-anthropic/token"},
							{Name: "url", Kind: providerservice.ProviderParamKindURL, Value: "https://api.anthropic.com"},
						},
					},
				},
			},
		}
		root := &cobra.Command{}
		var out bytes.Buffer
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}
		output := out.String()
		for _, want := range []string{`"items"`, `"my-anthropic"`, `"anthropic"`, `"token"`, `"my-anthropic/token"`, `"url"`} {
			if !strings.Contains(output, want) {
				t.Errorf("expected %q in JSON output, got: %s", want, output)
			}
		}
	})

	t.Run("json output empty list returns items array", func(t *testing.T) {
		t.Parallel()

		c := &providerListCmd{output: "json", store: &fakeProviderStore{}}
		root := &cobra.Command{}
		var out bytes.Buffer
		root.SetOut(&out)
		child := &cobra.Command{RunE: c.run}
		root.AddCommand(child)

		if err := child.RunE(child, []string{}); err != nil {
			t.Fatalf("run() failed: %v", err)
		}
		if !strings.Contains(out.String(), `"items"`) {
			t.Errorf("expected JSON with items key, got: %s", out.String())
		}
	})

	t.Run("store error propagates", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("store error")
		c := &providerListCmd{store: &fakeProviderStore{listErr: sentinel}}
		cmd := &cobra.Command{}
		var out bytes.Buffer
		cmd.SetOut(&out)
		err := c.run(cmd, []string{})
		if err == nil {
			t.Fatal("expected error when store fails")
		}
		if !errors.Is(err, sentinel) {
			t.Errorf("expected error to wrap sentinel, got: %v", err)
		}
	})
}
