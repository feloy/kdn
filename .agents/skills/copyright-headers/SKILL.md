---
name: copyright-headers
description: Add or update Apache License 2.0 copyright headers to Go source files, configuration files (YAML, TOML), shell scripts, and Makefiles
argument-hint: "[files to examine]"
---

# Copyright Headers

Add copyright headers to Go source files and configuration files.

## Description

This skill automatically adds Apache License 2.0 copyright headers to files. It supports:
- Go source files (`.go`)
- Configuration files (`.yaml`, `.yml`, `.toml`)
- Workflow files (`.github/workflows/*.yml`, `.github/workflows/*.yaml`)
- Shell scripts (`.sh`)
- Makefiles

## Instructions

When this skill is invoked:

1. **Identify target files**:
   - If the user specifies files or patterns, use those
   - Otherwise, scan the current directory and subdirectories for Go source files and configuration files
   - Exclude: vendor directories, `.git` directory, `go.mod`, `go.sum`

2. **Check for existing headers and update years**:
   - Read each file and check if it already has a copyright header in the first 20 lines
   - Look for keywords: "Copyright", "License", "Licensed", "SPDX-License-Identifier"
   - If a copyright header exists, check the year format and update if needed:
     - **Single past year** (e.g., `Copyright (C) 2024`) → Update to range: `Copyright (C) 2024-2026`
     - **Range ending before current year** (e.g., `Copyright (C) 2024-2025`) → Update end year: `Copyright (C) 2024-2026`
     - **Range with same start and end** (e.g., `Copyright (C) 2024-2024`) → Update to range: `Copyright (C) 2024-2026`
     - **Current year** (e.g., `Copyright (C) 2026`) → No change needed
     - **Range ending with current year** (e.g., `Copyright (C) 2024-2026`) → No change needed
     - **Future years or invalid ranges** → Leave as-is and optionally warn
   - If no copyright header exists, add one

3. **Add the copyright header based on file type**:

### For Go files (.go):
```go
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

```
- Place BEFORE the package declaration
- Add a blank line after the header

### For YAML, Shell, Makefiles, and other comment-based files:
```yaml
# Copyright (C) 2026 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0

```

4. **Special handling**:
   - **Shebangs**: For shell scripts with `#!/...`, place the copyright header AFTER the shebang, with a blank line between
   - **YAML directives**: For YAML files with `---` at the top, place the copyright header BEFORE the `---`

5. **Report results**:
   - List files that had headers added (new headers)
   - List files that had years updated (existing headers)
   - List files that were skipped (already up-to-date)
   - Show total count of files processed
   - Report any errors or warnings

## Usage Examples

Add headers to all Go files and config files:
```
/copyright-headers
```

Add headers to specific files:
```
/copyright-headers main.go root.go
```

Add headers to all Go files in a directory:
```
/copyright-headers pkg/cmd/**/*.go
```

## Important Notes

- Use the current year (2026) in the copyright notice
- Preserve existing file structure (shebangs, package declarations, etc.)
- Update copyright years in existing headers if they are outdated
- Skip generated files and vendor directories
- When updating years, preserve the original start year and update only the end year
