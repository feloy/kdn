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

// Package provider provides interfaces and implementations for managing user provider instances.
// Provider metadata (name, type, non-sensitive param values, and keychain reference names for
// sensitive params) is persisted in a JSON file under the kdn storage directory.
// Sensitive param values are stored in the system keychain.
package provider

import (
	"errors"

	"github.com/openkaiden/kdn/pkg/providerservice"
)

var (
	// ErrProviderAlreadyExists is returned when a provider with the same name already exists.
	ErrProviderAlreadyExists = errors.New("provider already exists")
	// ErrProviderNotFound is returned when no provider with the given name exists.
	ErrProviderNotFound = errors.New("provider not found")
)

// ProviderParamEntry holds a single parameter with its kind and value.
// For Kind=secret: Value is the keychain reference name (e.g. "my-anthropic/token");
// the actual secret is stored in the system keychain under that reference.
// For all other kinds: Value is the literal parameter value.
type ProviderParamEntry struct {
	Name       string
	Kind       providerservice.ProviderParamKind
	Value      string
	SecretType string // secret service type used when storing this param; only set for Kind=secret
}

// CreateParams holds all parameters needed to create a provider.
// For secret-kind entries, Value is the actual secret value (e.g. the API token);
// the store generates the keychain reference name, stores the secret in the keychain,
// and writes the reference name to the JSON file.
type CreateParams struct {
	Name   string
	Type   string
	Params []ProviderParamEntry
}

// ListItem holds the metadata fields returned by List.
// Secret-kind entries have an empty Value (the actual value is in the keychain).
type ListItem struct {
	Name   string
	Type   string
	Params []ProviderParamEntry
}

// Store manages persistent storage of provider instances.
type Store interface {
	// Create stores secret param values in the system keychain and persists
	// all metadata to the storage directory.
	Create(params CreateParams) error
	// List returns the metadata for all stored providers.
	List() ([]ListItem, error)
	// Get returns the metadata and secret values for the named provider.
	// The returned map contains secret param values keyed by param name.
	// Returns ErrProviderNotFound if no provider with the given name exists.
	Get(name string) (ListItem, map[string]string, error)
	// Remove deletes all secret param values from the system keychain and removes
	// the provider metadata from the storage directory.
	Remove(name string) error
}
