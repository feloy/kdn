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
	"encoding/json"
	"fmt"
	"path/filepath"

	api "github.com/openkaiden/kdn-api/cli/go"
	"github.com/openkaiden/kdn/pkg/provider"
	"github.com/spf13/cobra"
)

type providerRemoveCmd struct {
	store  provider.Store
	output string
}

func (c *providerRemoveCmd) preRun(cmd *cobra.Command, args []string) error {
	if c.output != "" && c.output != "json" {
		return fmt.Errorf("unsupported output format: %s (supported: json)", c.output)
	}
	if c.output == "json" {
		cmd.SilenceErrors = true
	}

	storageDir, err := cmd.Flags().GetString("storage")
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to read --storage flag: %w", err))
	}
	absStorageDir, err := filepath.Abs(storageDir)
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to resolve storage directory path: %w", err))
	}
	c.store = provider.NewStore(absStorageDir)
	return nil
}

func (c *providerRemoveCmd) run(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := c.store.Remove(name); err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to remove provider: %w", err))
	}

	if c.output == "json" {
		return c.outputJSON(cmd, name)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Provider %q removed successfully\n", name)
	return nil
}

func (c *providerRemoveCmd) outputJSON(cmd *cobra.Command, name string) error {
	response := api.ProviderName{Name: name}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to marshal to JSON: %w", err))
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func NewProviderRemoveCmd() *cobra.Command {
	c := &providerRemoveCmd{}

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a provider connection",
		Long:    "Remove a configured LLM provider connection.",
		Example: `# Remove a provider connection by name
kdn provider remove my-anthropic

# Remove a provider connection with JSON output
kdn provider remove my-anthropic --output json

# Remove a provider connection with short flag
kdn provider remove my-anthropic -o json`,
		Args:    cobra.ExactArgs(1),
		PreRunE: c.preRun,
		RunE:    c.run,
	}

	cmd.Flags().StringVarP(&c.output, "output", "o", "", "Output format (supported: json)")
	cmd.RegisterFlagCompletionFunc("output", newOutputFlagCompletion([]string{"json"}))
	cmd.ValidArgsFunction = completeProviderName

	return cmd
}
