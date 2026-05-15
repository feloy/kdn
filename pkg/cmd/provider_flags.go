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
	"fmt"
	"strings"

	"github.com/openkaiden/kdn/pkg/provider"
	"github.com/openkaiden/kdn/pkg/providerservice"
	"github.com/spf13/cobra"
)

// registerProviderParamFlagCompletions registers shell completion for provider param flags.
// Credential-kind params (e.g. --credentials for vertexai) get file completion so the shell
// can suggest paths; all other param kinds get no completion.
func registerProviderParamFlagCompletions(cmd *cobra.Command, services []providerservice.ProviderService) {
	seen := make(map[string]bool)
	for _, svc := range services {
		for _, p := range svc.Params() {
			if seen[p.Name] {
				continue
			}
			seen[p.Name] = true
			if p.Kind == providerservice.ProviderParamKindCredential {
				cmd.RegisterFlagCompletionFunc(p.Name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
					return nil, cobra.ShellCompDirectiveDefault
				})
			}
		}
	}
}

// providerParamFlagInfo tracks a param definition together with all provider type names that use it.
type providerParamFlagInfo struct {
	param     providerservice.ProviderParam
	typeNames []string
}

// buildProviderParamFlagInfos returns an ordered list of unique param names and a map from
// param name to its flag info (definition + all type names that use it).
// The order follows the first appearance of each param across the services slice.
func buildProviderParamFlagInfos(services []providerservice.ProviderService) ([]string, map[string]*providerParamFlagInfo) {
	order := make([]string, 0)
	infos := make(map[string]*providerParamFlagInfo)
	for _, svc := range services {
		for _, p := range svc.Params() {
			if _, seen := infos[p.Name]; !seen {
				order = append(order, p.Name)
				infos[p.Name] = &providerParamFlagInfo{param: p}
			}
			infos[p.Name].typeNames = append(infos[p.Name].typeNames, svc.Name())
		}
	}
	return order, infos
}

// registerProviderParamFlags registers a string flag on cmd for each unique param across
// the given services. The flag description lists every provider type that uses the param.
func registerProviderParamFlags(cmd *cobra.Command, services []providerservice.ProviderService) {
	order, infos := buildProviderParamFlagInfos(services)
	for _, name := range order {
		info := infos[name]
		typeList := strings.Join(info.typeNames, ", ")
		cmd.Flags().String(info.param.Name, "", fmt.Sprintf("%s (for --type=%s)", info.param.Description, typeList))
	}
}

// allProviderParamNames returns the set of all unique param names across the given services.
func allProviderParamNames(services []providerservice.ProviderService) map[string]bool {
	names := make(map[string]bool)
	for _, svc := range services {
		for _, p := range svc.Params() {
			names[p.Name] = true
		}
	}
	return names
}

// collectProviderParams reads the provider params for the given service from the cobra command's
// flags, returning only those flags that were explicitly set by the user.
func collectProviderParams(cmd *cobra.Command, svc providerservice.ProviderService) ([]provider.ProviderParamEntry, error) {
	params := make([]provider.ProviderParamEntry, 0, len(svc.Params()))
	for _, p := range svc.Params() {
		if !cmd.Flags().Changed(p.Name) {
			continue
		}
		val, err := cmd.Flags().GetString(p.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to read --%s flag: %w", p.Name, err)
		}
		params = append(params, provider.ProviderParamEntry{
			Name:       p.Name,
			Kind:       p.Kind,
			Value:      val,
			SecretType: p.SecretType,
		})
	}
	return params, nil
}
