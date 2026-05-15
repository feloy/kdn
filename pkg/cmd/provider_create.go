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
	"sort"
	"strings"

	"github.com/openkaiden/kdn/pkg/provider"
	"github.com/openkaiden/kdn/pkg/providerservice"
	"github.com/openkaiden/kdn/pkg/providerservicesetup"
	"github.com/spf13/cobra"
)

type providerCreateCmd struct {
	providerType string
	output       string
	store        provider.Store
	validTypes   []string
	// params is populated in preRun from the flags explicitly set by the user.
	params []provider.ProviderParamEntry
}

func (c *providerCreateCmd) isValidType(t string) bool {
	for _, v := range c.validTypes {
		if t == v {
			return true
		}
	}
	return false
}

func (c *providerCreateCmd) preRun(cmd *cobra.Command, args []string) error {
	if c.output != "" && c.output != "json" {
		return fmt.Errorf("unsupported output format: %s (supported: json)", c.output)
	}
	if c.output == "json" {
		cmd.SilenceErrors = true
	}

	if c.providerType == "" {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("--type is required"))
	}
	if !c.isValidType(c.providerType) {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("invalid --type %q: must be one of %s", c.providerType, strings.Join(c.validTypes, ", ")))
	}

	services := providerservicesetup.ListServices()

	// Build the set of valid params for the selected type.
	validParams := make(map[string]providerservice.ProviderParam)
	var selectedSvc providerservice.ProviderService
	for _, s := range services {
		if s.Name() == c.providerType {
			selectedSvc = s
			for _, p := range s.Params() {
				validParams[p.Name] = p
			}
			break
		}
	}

	// Reject flags that don't belong to the selected type.
	for paramName := range allProviderParamNames(services) {
		if cmd.Flags().Changed(paramName) {
			if _, ok := validParams[paramName]; !ok {
				return outputErrorIfJSON(cmd, c.output, fmt.Errorf("--%s is not valid for --type=%s", paramName, c.providerType))
			}
		}
	}

	// Verify all required params were explicitly provided.
	for _, p := range selectedSvc.Params() {
		if p.Required && !cmd.Flags().Changed(p.Name) {
			return outputErrorIfJSON(cmd, c.output, fmt.Errorf("--%s is required for --type=%s", p.Name, c.providerType))
		}
	}

	// Collect param entries from the flags that were explicitly set.
	params, err := collectProviderParams(cmd, selectedSvc)
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, err)
	}
	c.params = params

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

func (c *providerCreateCmd) run(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := c.store.Create(provider.CreateParams{
		Name:   name,
		Type:   c.providerType,
		Params: c.params,
	}); err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to create provider: %w", err))
	}

	if c.output == "json" {
		return c.outputJSON(cmd, name)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Provider %q created successfully\n", name)
	return nil
}

func (c *providerCreateCmd) outputJSON(cmd *cobra.Command, name string) error {
	response := providerNameOutput{Name: name}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return outputErrorIfJSON(cmd, c.output, fmt.Errorf("failed to marshal to JSON: %w", err))
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

// paramPlaceholder returns the usage placeholder string for a provider param.
// Secrets show <value> to avoid hinting at the param name; credential params
// append -path to signal that a file path is expected; all others use <paramname>.
func paramPlaceholder(p providerservice.ProviderParam) string {
	switch p.Kind {
	case providerservice.ProviderParamKindSecret:
		return "<secret>"
	case providerservice.ProviderParamKindCredential:
		return "<" + p.Name + "-path>"
	default:
		return "<" + p.Name + ">"
	}
}

// buildProviderCreateUsage builds the synopsis section of the Long description dynamically
// from the registered provider services. Types are listed alphabetically; within each type,
// required params come first (alphabetical), followed by optional params (alphabetical).
func buildProviderCreateUsage(services []providerservice.ProviderService) string {
	sorted := make([]providerservice.ProviderService, len(services))
	copy(sorted, services)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name() < sorted[j].Name() })

	lines := make([]string, 0, len(sorted))
	for _, svc := range sorted {
		var required, optional []providerservice.ProviderParam
		for _, p := range svc.Params() {
			if p.Required {
				required = append(required, p)
			} else {
				optional = append(optional, p)
			}
		}
		sort.Slice(required, func(i, j int) bool { return required[i].Name < required[j].Name })
		sort.Slice(optional, func(i, j int) bool { return optional[i].Name < optional[j].Name })

		parts := []string{"--type " + svc.Name()}
		for _, p := range required {
			parts = append(parts, "--"+p.Name+" "+paramPlaceholder(p))
		}
		for _, p := range optional {
			parts = append(parts, "[--"+p.Name+" "+paramPlaceholder(p)+"]")
		}
		lines = append(lines, strings.Join(parts, " "))
	}

	return "create <name>\n  " + strings.Join(lines, " |\n  ")
}

func NewProviderCreateCmd() *cobra.Command {
	availableTypes := providerservicesetup.ListAvailable()
	sort.Strings(availableTypes)
	typesStr := strings.Join(availableTypes, ", ")

	c := &providerCreateCmd{validTypes: availableTypes}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new provider connection",
		Long: "Create a new LLM provider connection.\n\nUsage:\n  " +
			buildProviderCreateUsage(providerservicesetup.ListServices()),
		Example: `# Create an Anthropic provider connection
kdn provider create my-anthropic --type anthropic --token my-secret-token

# Create a Vertex AI provider connection
kdn provider create my-vertexai --type vertexai --project my-project --region us-central1 --credentials ~/.config/gcloud/application_default_credentials.json

# Create an Anthropic provider connection with JSON output
kdn provider create my-anthropic --type anthropic --token my-secret-token --output json`,
		Args:    cobra.ExactArgs(1),
		PreRunE: c.preRun,
		RunE:    c.run,
	}

	cmd.Flags().StringVar(&c.providerType, "type", "", fmt.Sprintf("Type of provider (%s)", typesStr))
	cmd.RegisterFlagCompletionFunc("type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return availableTypes, cobra.ShellCompDirectiveNoFileComp
	})

	registerProviderParamFlags(cmd, providerservicesetup.ListServices())
	registerProviderParamFlagCompletions(cmd, providerservicesetup.ListServices())

	cmd.Flags().StringVarP(&c.output, "output", "o", "", "Output format (supported: json)")
	cmd.RegisterFlagCompletionFunc("output", newOutputFlagCompletion([]string{"json"}))

	return cmd
}
