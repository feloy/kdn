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

package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/openkaiden/kdn/pkg/providerservice"
	"github.com/openkaiden/kdn/pkg/secret"
)

// fakeSecretStore is a fake implementation of secret.Store for use in tests.
type fakeSecretStore struct {
	secrets      map[string]string
	descriptions map[string]string
	types        map[string]string
	createErr    error
	removeErr    error
	getErr       error
}

func newFakeSecretStore() *fakeSecretStore {
	return &fakeSecretStore{
		secrets:      make(map[string]string),
		descriptions: make(map[string]string),
		types:        make(map[string]string),
	}
}

func (f *fakeSecretStore) Create(params secret.CreateParams) error {
	if f.createErr != nil {
		return f.createErr
	}
	if _, exists := f.secrets[params.Name]; exists {
		return fmt.Errorf("secret %q: %w", params.Name, secret.ErrSecretAlreadyExists)
	}
	f.secrets[params.Name] = params.Value
	f.descriptions[params.Name] = params.Description
	f.types[params.Name] = params.Type
	return nil
}

func (f *fakeSecretStore) List() ([]secret.ListItem, error) {
	return nil, nil
}

func (f *fakeSecretStore) Get(name string) (secret.ListItem, string, error) {
	if f.getErr != nil {
		return secret.ListItem{}, "", f.getErr
	}
	val, ok := f.secrets[name]
	if !ok {
		return secret.ListItem{}, "", fmt.Errorf("secret %q: %w", name, secret.ErrSecretNotFound)
	}
	return secret.ListItem{Name: name}, val, nil
}

func (f *fakeSecretStore) Remove(name string) error {
	if f.removeErr != nil {
		return f.removeErr
	}
	if _, ok := f.secrets[name]; !ok {
		return fmt.Errorf("secret %q: %w", name, secret.ErrSecretNotFound)
	}
	delete(f.secrets, name)
	return nil
}

var _ secret.Store = (*fakeSecretStore)(nil)

func TestStore_Create_StoresSecretParamViaSecretStore(t *testing.T) {
	t.Parallel()

	ss := newFakeSecretStore()
	st := newStoreWithSecretStore(t.TempDir(), ss)

	err := st.Create(CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-ant-secret", SecretType: "anthropic"},
		},
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Secret should be stored under "{providerName}/{paramName}"
	val, ok := ss.secrets["my-anthropic/token"]
	if !ok {
		t.Fatal("expected secret to be stored in secret store as 'my-anthropic/token'")
	}
	if val != "sk-ant-secret" {
		t.Errorf("secret value = %q, want %q", val, "sk-ant-secret")
	}
	// Type must match the param's SecretType, not a hardcoded value.
	if ss.types["my-anthropic/token"] != "anthropic" {
		t.Errorf("type = %q, want %q", ss.types["my-anthropic/token"], "anthropic")
	}
	// Description should identify the param and provider.
	if ss.descriptions["my-anthropic/token"] != "token for my-anthropic provider" {
		t.Errorf("description = %q, want %q", ss.descriptions["my-anthropic/token"], "token for my-anthropic provider")
	}
}

func TestStore_Create_SavesMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	st := newStoreWithSecretStore(dir, newFakeSecretStore())

	err := st.Create(CreateParams{
		Name: "my-vertexai",
		Type: "vertexai",
		Params: []ProviderParamEntry{
			{Name: "project", Kind: providerservice.ProviderParamKindText, Value: "my-project"},
			{Name: "region", Kind: providerservice.ProviderParamKindText, Value: "us-central1"},
		},
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, providersFileName))
	if err != nil {
		t.Fatalf("failed to read providers file: %v", err)
	}

	var pf providersFile
	if err := json.Unmarshal(data, &pf); err != nil {
		t.Fatalf("failed to parse providers file: %v", err)
	}

	if len(pf.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(pf.Providers))
	}

	rec := pf.Providers[0]
	if rec.Name != "my-vertexai" {
		t.Errorf("Name: want %q, got %q", "my-vertexai", rec.Name)
	}
	if rec.Type != "vertexai" {
		t.Errorf("Type: want %q, got %q", "vertexai", rec.Type)
	}
	if len(rec.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(rec.Params))
	}
	if rec.Params[0].Value != "my-project" {
		t.Errorf("project value: want %q, got %q", "my-project", rec.Params[0].Value)
	}
	if rec.Params[1].Value != "us-central1" {
		t.Errorf("region value: want %q, got %q", "us-central1", rec.Params[1].Value)
	}
}

func TestStore_Create_StoresSecretRefInJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	st := newStoreWithSecretStore(dir, newFakeSecretStore())

	err := st.Create(CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-secret", SecretType: "anthropic"},
		},
	})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, providersFileName))
	var pf providersFile
	_ = json.Unmarshal(data, &pf)

	// The JSON should store the reference name, not the actual secret value.
	if pf.Providers[0].Params[0].Value != "my-anthropic/token" {
		t.Errorf("expected ref name %q in JSON, got %q", "my-anthropic/token", pf.Providers[0].Params[0].Value)
	}
}

func TestStore_Create_ErrorsOnDuplicate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ss := newFakeSecretStore()
	st := newStoreWithSecretStore(dir, ss)

	params := CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "v1", SecretType: "anthropic"},
		},
	}
	if err := st.Create(params); err != nil {
		t.Fatalf("first Create() failed: %v", err)
	}

	secretsBefore := len(ss.secrets)
	params.Params[0].Value = "v2"
	err := st.Create(params)
	if err == nil {
		t.Fatal("expected error when creating duplicate provider")
	}
	if !errors.Is(err, ErrProviderAlreadyExists) {
		t.Errorf("expected ErrProviderAlreadyExists, got: %v", err)
	}
	// Secret store must not be touched when the duplicate is detected.
	if len(ss.secrets) != secretsBefore {
		t.Errorf("secret store was modified despite duplicate: got %d entries, want %d", len(ss.secrets), secretsBefore)
	}
}

func TestStore_List(t *testing.T) {
	t.Parallel()

	t.Run("empty when no providers exist", func(t *testing.T) {
		t.Parallel()

		st := newStoreWithSecretStore(t.TempDir(), newFakeSecretStore())
		items, err := st.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("expected 0 items, got %d", len(items))
		}
	})

	t.Run("returns metadata sorted by name", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		st := newStoreWithSecretStore(dir, newFakeSecretStore())

		if err := st.Create(CreateParams{
			Name: "z-provider",
			Type: "vertexai",
			Params: []ProviderParamEntry{
				{Name: "project", Kind: providerservice.ProviderParamKindText, Value: "proj-z"},
			},
		}); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		if err := st.Create(CreateParams{
			Name: "a-provider",
			Type: "anthropic",
			Params: []ProviderParamEntry{
				{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "tok", SecretType: "anthropic"},
			},
		}); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		items, err := st.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		// Items must be sorted by name.
		if items[0].Name != "a-provider" {
			t.Errorf("items[0].Name = %q, want %q", items[0].Name, "a-provider")
		}
		if items[1].Name != "z-provider" {
			t.Errorf("items[1].Name = %q, want %q", items[1].Name, "z-provider")
		}
	})

	t.Run("secret param value is the reference name in list", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		st := newStoreWithSecretStore(dir, newFakeSecretStore())

		if err := st.Create(CreateParams{
			Name: "my-anthropic",
			Type: "anthropic",
			Params: []ProviderParamEntry{
				{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-secret", SecretType: "anthropic"},
			},
		}); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		items, _ := st.List()
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		// Value for secret params is the secret reference name, not the actual secret value.
		if items[0].Params[0].Value != "my-anthropic/token" {
			t.Errorf("expected reference name %q for secret param in list, got %q", "my-anthropic/token", items[0].Params[0].Value)
		}
	})
}

func TestStore_Get_ReturnsMetadataAndSecrets(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	st := newStoreWithSecretStore(dir, newFakeSecretStore())

	if err := st.Create(CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-ant-secret", SecretType: "anthropic"},
			{Name: "url", Kind: providerservice.ProviderParamKindURL, Value: "https://api.anthropic.com"},
		},
	}); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	item, secrets, err := st.Get("my-anthropic")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if item.Name != "my-anthropic" {
		t.Errorf("item.Name = %q, want %q", item.Name, "my-anthropic")
	}
	if item.Type != "anthropic" {
		t.Errorf("item.Type = %q, want %q", item.Type, "anthropic")
	}
	if secrets["token"] != "sk-ant-secret" {
		t.Errorf("secrets[token] = %q, want %q", secrets["token"], "sk-ant-secret")
	}
	// Non-secret param should not appear in secrets map.
	if _, ok := secrets["url"]; ok {
		t.Error("expected 'url' not to appear in secrets map")
	}
	// URL param value should be in item.Params.
	var urlVal string
	for _, p := range item.Params {
		if p.Name == "url" {
			urlVal = p.Value
		}
	}
	if urlVal != "https://api.anthropic.com" {
		t.Errorf("url param value = %q, want %q", urlVal, "https://api.anthropic.com")
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	t.Parallel()

	st := newStoreWithSecretStore(t.TempDir(), newFakeSecretStore())

	_, _, err := st.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error when provider does not exist")
	}
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got: %v", err)
	}
}

func TestStore_Remove_DeletesSecretsAndMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ss := newFakeSecretStore()
	st := newStoreWithSecretStore(dir, ss)

	if err := st.Create(CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-secret", SecretType: "anthropic"},
		},
	}); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if err := st.Remove("my-anthropic"); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Secret should be gone from secret store.
	if _, ok := ss.secrets["my-anthropic/token"]; ok {
		t.Error("expected secret to be removed from secret store")
	}

	// Provider should be gone from JSON.
	data, err := os.ReadFile(filepath.Join(dir, providersFileName))
	if err != nil {
		t.Fatalf("failed to read providers file: %v", err)
	}
	var pf providersFile
	_ = json.Unmarshal(data, &pf)
	if len(pf.Providers) != 0 {
		t.Errorf("expected 0 providers after Remove, got %d", len(pf.Providers))
	}
}

func TestStore_Remove_NotFound(t *testing.T) {
	t.Parallel()

	st := newStoreWithSecretStore(t.TempDir(), newFakeSecretStore())

	err := st.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error when provider does not exist")
	}
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got: %v", err)
	}
}

func TestStore_Remove_MissingSecretStillRemovesMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ss := newFakeSecretStore()
	st := newStoreWithSecretStore(dir, ss)

	if err := st.Create(CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-secret", SecretType: "anthropic"},
		},
	}); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Manually delete the secret from the secret store to simulate out-of-sync state.
	delete(ss.secrets, "my-anthropic/token")

	// Remove should succeed even when the secret is already gone.
	if err := st.Remove("my-anthropic"); err != nil {
		t.Fatalf("Remove() should succeed even when secret is missing: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, providersFileName))
	if err != nil {
		t.Fatalf("failed to read providers file: %v", err)
	}
	var pf providersFile
	_ = json.Unmarshal(data, &pf)
	if len(pf.Providers) != 0 {
		t.Errorf("expected 0 providers after Remove, got %d", len(pf.Providers))
	}
}

func TestStore_Create_SecretStoreError(t *testing.T) {
	t.Parallel()

	ss := newFakeSecretStore()
	ss.createErr = os.ErrPermission
	st := newStoreWithSecretStore(t.TempDir(), ss)

	err := st.Create(CreateParams{
		Name: "my-anthropic",
		Type: "anthropic",
		Params: []ProviderParamEntry{
			{Name: "token", Kind: providerservice.ProviderParamKindSecret, Value: "sk-secret", SecretType: "anthropic"},
		},
	})
	if err == nil {
		t.Fatal("expected error when secret store fails")
	}
}
