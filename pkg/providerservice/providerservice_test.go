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

package providerservice

import (
	"testing"
)

func TestProviderParamKind_Constants(t *testing.T) {
	t.Parallel()

	if ProviderParamKindSecret != "secret" {
		t.Errorf("ProviderParamKindSecret = %q, want %q", ProviderParamKindSecret, "secret")
	}
	if ProviderParamKindCredential != "credential" {
		t.Errorf("ProviderParamKindCredential = %q, want %q", ProviderParamKindCredential, "credential")
	}
	if ProviderParamKindURL != "url" {
		t.Errorf("ProviderParamKindURL = %q, want %q", ProviderParamKindURL, "url")
	}
	if ProviderParamKindText != "text" {
		t.Errorf("ProviderParamKindText = %q, want %q", ProviderParamKindText, "text")
	}
}

func TestProviderParam_Fields(t *testing.T) {
	t.Parallel()

	p := ProviderParam{
		Name:        "token",
		Description: "API token",
		Required:    true,
		Kind:        ProviderParamKindSecret,
	}

	if p.Name != "token" {
		t.Errorf("Name = %q, want %q", p.Name, "token")
	}
	if p.Description != "API token" {
		t.Errorf("Description = %q, want %q", p.Description, "API token")
	}
	if !p.Required {
		t.Error("Required = false, want true")
	}
	if p.Kind != ProviderParamKindSecret {
		t.Errorf("Kind = %q, want %q", p.Kind, ProviderParamKindSecret)
	}
}

func TestNewProviderService(t *testing.T) {
	t.Parallel()

	params := []ProviderParam{
		{Name: "token", Description: "API token", Required: true, Kind: ProviderParamKindSecret},
		{Name: "url", Description: "Custom base URL", Required: false, Kind: ProviderParamKindURL},
	}
	svc := NewProviderService("anthropic", "Anthropic Claude AI models", params)

	if svc == nil {
		t.Fatal("NewProviderService() returned nil")
	}
	if svc.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", svc.Name(), "anthropic")
	}
	if svc.Description() != "Anthropic Claude AI models" {
		t.Errorf("Description() = %q, want %q", svc.Description(), "Anthropic Claude AI models")
	}
	if len(svc.Params()) != 2 {
		t.Fatalf("Params() returned %d params, want 2", len(svc.Params()))
	}
	if svc.Params()[0].Name != "token" || svc.Params()[0].Kind != ProviderParamKindSecret || !svc.Params()[0].Required {
		t.Errorf("Params()[0] = %+v, unexpected", svc.Params()[0])
	}
	if svc.Params()[1].Name != "url" || svc.Params()[1].Kind != ProviderParamKindURL || svc.Params()[1].Required {
		t.Errorf("Params()[1] = %+v, unexpected", svc.Params()[1])
	}
}

func TestNewProviderService_EmptyParams(t *testing.T) {
	t.Parallel()

	svc := NewProviderService("minimal", "", nil)
	if svc.Name() != "minimal" {
		t.Errorf("Name() = %q, want %q", svc.Name(), "minimal")
	}
	if svc.Description() != "" {
		t.Errorf("Description() = %q, want empty", svc.Description())
	}
	if svc.Params() != nil {
		t.Errorf("Params() = %v, want nil", svc.Params())
	}
}
