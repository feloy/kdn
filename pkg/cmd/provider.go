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

	"github.com/spf13/cobra"
)

// providerNameOutput is the JSON output for provider create and remove commands.
type providerNameOutput struct {
	Name string `json:"name"`
}

// providerParamOutput is the JSON output for a single provider parameter.
type providerParamOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// providerInfoOutput is the JSON output for a single provider entry.
type providerInfoOutput struct {
	Name   string                `json:"name"`
	Type   string                `json:"type"`
	Params []providerParamOutput `json:"params,omitempty"`
}

// providersListOutput is the JSON output for the provider list command.
type providersListOutput struct {
	Items []providerInfoOutput `json:"items"`
}

func NewProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage LLM provider connections",
		Long:  "Manage LLM provider connections",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewProviderCreateCmd())
	cmd.AddCommand(NewProviderListCmd())
	cmd.AddCommand(NewProviderRemoveCmd())

	return cmd
}
