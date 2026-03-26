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

import (
	"context"
	"io"
	"sync"

	"github.com/yarlson/pin"
)

// textLogger is an implementation of StepLogger that outputs to a writer with spinners.
type textLogger struct {
	writer         io.Writer
	mu             sync.Mutex
	currentSpinner *pin.Pin
	currentCancel  context.CancelFunc
	completedMsg   string
	failed         bool
}

// Compile-time check to ensure textLogger implements StepLogger interface
var _ StepLogger = (*textLogger)(nil)

// NewTextLogger creates a new text step logger that outputs to the given writer.
func NewTextLogger(w io.Writer) StepLogger {
	return &textLogger{
		writer: w,
	}
}

// Start begins a new step with the given messages.
// Automatically completes the previous step with its completion message if one exists.
func (t *textLogger) Start(inProgress, completed string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Auto-complete previous step if exists and wasn't failed
	if t.currentSpinner != nil {
		if !t.failed {
			t.currentSpinner.Stop(t.completedMsg)
		}
		// Always cancel to clean up goroutines
		if t.currentCancel != nil {
			t.currentCancel()
		}
	}

	// Reset failed state
	t.failed = false

	// Create and start new spinner with our writer
	t.currentSpinner = pin.New(inProgress, pin.WithWriter(t.writer))
	t.completedMsg = completed
	t.currentCancel = t.currentSpinner.Start(context.Background())
}

// Complete marks the current step as successfully completed using its completion message.
func (t *textLogger) Complete() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.currentSpinner != nil && !t.failed {
		t.currentSpinner.Stop(t.completedMsg)
		if t.currentCancel != nil {
			t.currentCancel()
		}
		t.currentSpinner = nil
		t.currentCancel = nil
	}
}

// Fail marks the current step as failed with the given error.
func (t *textLogger) Fail(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.currentSpinner != nil {
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		t.currentSpinner.Fail(errMsg)
		if t.currentCancel != nil {
			t.currentCancel()
		}
		t.failed = true
		t.currentSpinner = nil
		t.currentCancel = nil
	}
}
