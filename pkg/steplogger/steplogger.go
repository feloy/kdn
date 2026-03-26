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

// Package steplogger provides interfaces and implementations for logging
// operational steps during runtime operations.
package steplogger

// StepLogger provides an interface for logging operational steps.
// Implementations can display steps differently based on output mode (text vs JSON).
//
// The Start() method takes two messages: one for in-progress state and one for completion.
// It automatically completes the previous step with its completion message before starting a new one.
// The Complete() method should be called with defer to complete the last step.
// The Fail() method marks the current step as failed.
type StepLogger interface {
	// Start begins a new step with the given messages.
	// The inProgress message is shown while the step is running.
	// The completed message is shown when the step completes successfully.
	// Automatically completes the previous step with its completion message if one exists.
	Start(inProgress, completed string)

	// Complete marks the current step as successfully completed using its completion message.
	// Typically called with defer to complete the last step.
	Complete()

	// Fail marks the current step as failed with the given error.
	// After failure, the step logger will not auto-complete on the next Start().
	Fail(err error)
}
