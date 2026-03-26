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

package steplogger

// noopLogger is a no-op implementation of StepLogger.
// Used for tests, backward compatibility, and JSON mode.
type noopLogger struct{}

// Compile-time check to ensure noopLogger implements StepLogger interface
var _ StepLogger = (*noopLogger)(nil)

// NewNoOpLogger creates a new no-op step logger.
func NewNoOpLogger() StepLogger {
	return &noopLogger{}
}

// Start is a no-op.
func (n *noopLogger) Start(inProgress, completed string) {}

// Complete is a no-op.
func (n *noopLogger) Complete() {}

// Fail is a no-op.
func (n *noopLogger) Fail(err error) {}
