---
name: working-with-kdn-api
description: Guide to updating the kdn-api Go modules, what to check afterwards, and how to keep schemas and the config merger in sync
argument-hint: ""
---

# Working with kdn-api

The `openkaiden/kdn-api` repository defines the OpenAPI specs and generated Go types that kdn depends on. This skill covers how to bump the modules, and what must be reviewed or updated afterwards.

## Two Go modules

kdn imports two Go modules from kdn-api:

| Module path | Package alias | Purpose |
|---|---|---|
| `github.com/openkaiden/kdn-api/workspace-configuration/go` | `workspace` | `WorkspaceConfiguration` and all its sub-types (`Mount`, `EnvironmentVariable`, `McpConfiguration`, …) |
| `github.com/openkaiden/kdn-api/cli/go` | `api` | CLI output types (`Workspace`, `WorkspaceId`, `WorkspacePaths`, …) used in `pkg/cmd/conversion.go` |

Both modules are versioned together (same `vX.Y.Z` tag). Both must always be bumped at the same time.

## Bumping the modules

```bash
go get github.com/openkaiden/kdn-api/workspace-configuration/go@vX.Y.Z
go get github.com/openkaiden/kdn-api/cli/go@vX.Y.Z
go mod tidy
```

Verify the bump compiled and tests pass:

```bash
make build
make test
```

### Commit message convention

```text
chore(deps): bump kdn-api modules to vX.Y.Z
```

## Checklist after bumping

Always work through this list after any kdn-api bump, regardless of how small the diff looks.

### 1. Check the kdn-api release notes

Read the GitHub release for the new tag to understand what changed:

```bash
gh api repos/openkaiden/kdn-api/releases/tags/vX.Y.Z --jq '.body'
```

### 2. If `WorkspaceConfiguration` or its sub-types changed

`WorkspaceConfiguration` is defined in `workspace-configuration/openapi.yaml` and code-generated into `workspace-configuration/go/workspace.go`. When it gains new fields, renames fields, or changes field types:

**a. Update `pkg/config/merger.go`**

`Merge()` and `copyConfig()` must handle every field of `WorkspaceConfiguration` explicitly. Fields not wired in are silently dropped — there is no compile-time check. Add the new field to both functions following the pattern of existing fields. `mergeFeatures()` / `copyFeatures()` and `mergePorts()` are examples of how compound fields are handled.

**b. Update all three JSON schemas in `schemas/`**

The schemas are maintained manually, derived from the OpenAPI spec. All three files must stay in sync because they share the same `WorkspaceConfiguration` sub-schema:

- `schemas/workspace.json` — per-workspace config (`.kaiden/workspace.json`)
- `schemas/agents.json` — per-agent config (`~/.kdn/config/agents.json`)
- `schemas/projects.json` — per-project config (`~/.kdn/config/projects.json`)

The `$defs` section in each file is identical. Update all three at once. The source of truth is the OpenAPI spec:

```bash
gh api repos/openkaiden/kdn-api/contents/workspace-configuration/openapi.yaml --jq '.content' | base64 -d
```

**c. Check `pkg/config/config.go` validation**

The `validate()` function may need updating if new fields introduce new constraints.

**d. Check `pkg/config/workspaceupdater.go` and `projectsupdater.go`**

These updaters manipulate specific fields. New fields exposed to users may warrant new updater methods.

### 3. If CLI types (`api.*`) changed

`api.*` types are used in `pkg/cmd/conversion.go`. Check that `instanceToWorkspace()` and related functions still produce valid output for any renamed or added fields:

```bash
grep -n "api\." pkg/cmd/conversion.go
```

### 4. If `WorkspaceConfiguration` type changes affect the config merger tests

`pkg/config/merger_test.go` tests `Merge()` exhaustively. New fields need corresponding test cases.

## Where `workspace.WorkspaceConfiguration` is used

Every package that reads or writes workspace config imports `workspace-configuration/go`. To audit impact of a type change:

```bash
grep -rn "workspace\." pkg/ --include="*.go" | grep -v "_test.go" | grep -v "^Binary"
```

Key locations:
- `pkg/config/` — loading, merging, and updating config files
- `pkg/runtime/podman/create.go` — consuming config when creating containers
- `pkg/runtime/openshell/create.go` — consuming config for OpenShell workspaces
- `pkg/agent/*.go` — reading skills, MCP, and environment from config
- `pkg/autoconf/` — writing config entries discovered from the environment
