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
	"strings"

	"github.com/fatih/color"
	api "github.com/openkaiden/kdn-api/cli/go"
	"github.com/openkaiden/kdn/pkg/provider"
	"github.com/openkaiden/kdn/pkg/providerservice"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

type providerListCmd struct {
	store  provider.Store
	output string
}

func (c *providerListCmd) preRun(cmd *cobra.Command, args []string) error {
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

func (c *providerListCmd) run(cmd *cobra.Command, args []string) error {
	items, err := c.store.List()
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to list providers: %w", err))
	}

	if c.output == "json" {
		return c.outputJSON(cmd, items)
	}

	return c.displayTable(cmd, items)
}

func (c *providerListCmd) displayTable(cmd *cobra.Command, items []provider.ListItem) error {
	out := cmd.OutOrStdout()
	if len(items) == 0 {
		fmt.Fprintln(out, "No providers found")
		return nil
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("NAME", "TYPE", "PARAMS")
	tbl.WithWriter(out)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, item := range items {
		tbl.AddRow(item.Name, item.Type, formatProviderParams(item))
	}

	tbl.Print()
	return nil
}

// formatProviderParams formats params as "key=value, key=value".
// Secret-kind params are shown as "key=<secret>" to avoid displaying the reference name.
func formatProviderParams(item provider.ListItem) string {
	parts := make([]string, 0, len(item.Params))
	for _, p := range item.Params {
		if p.Kind == providerservice.ProviderParamKindSecret {
			parts = append(parts, fmt.Sprintf("%s=<secret>", p.Name))
		} else if p.Value != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", p.Name, p.Value))
		}
	}
	return strings.Join(parts, ", ")
}

func (c *providerListCmd) outputJSON(cmd *cobra.Command, items []provider.ListItem) error {
	output := api.ProvidersList{
		Items: make([]api.ProviderInfo, 0, len(items)),
	}
	for _, item := range items {
		info := api.ProviderInfo{
			Name:   item.Name,
			Type:   item.Type,
			Params: make([]api.ProviderParam, 0, len(item.Params)),
		}
		for _, p := range item.Params {
			info.Params = append(info.Params, api.ProviderParam{
				Name:  p.Name,
				Value: p.Value,
			})
		}
		output.Items = append(output.Items, info)
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to marshal providers to JSON: %w", err))
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func NewProviderListCmd() *cobra.Command {
	c := &providerListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all provider connections",
		Long:  "List all configured LLM provider connections",
		Example: `# List all provider connections
kdn provider list

# List provider connections in JSON format
kdn provider list --output json

# List provider connections using short flag
kdn provider list -o json`,
		Args:    cobra.NoArgs,
		PreRunE: c.preRun,
		RunE:    c.run,
	}

	cmd.Flags().StringVarP(&c.output, "output", "o", "", "Output format (supported: json)")
	cmd.RegisterFlagCompletionFunc("output", newOutputFlagCompletion([]string{"json"}))

	return cmd
}
