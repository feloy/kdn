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
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNoOpLogger(t *testing.T) {
	t.Parallel()

	logger := NewNoOpLogger()

	// Should not panic on any calls
	logger.Start("test", "test complete")
	logger.Complete()
	logger.Fail(errors.New("test error"))
	logger.Start("test2", "test2 complete")
	logger.Complete()
}

func TestTextLogger_Basic(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start a step
	logger.Start("Step 1", "Step 1 done")
	time.Sleep(100 * time.Millisecond) // Let spinner run briefly

	// Complete the step
	logger.Complete()

	// Output should contain the completion message
	output := buf.String()
	if !strings.Contains(output, "Step 1 done") {
		t.Errorf("Expected output to contain 'Step 1 done', got: %s", output)
	}
}

func TestTextLogger_AutoComplete(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start first step
	logger.Start("Step 1", "Step 1 done")
	time.Sleep(50 * time.Millisecond)

	// Start second step (should auto-complete first)
	logger.Start("Step 2", "Step 2 done")
	time.Sleep(50 * time.Millisecond)

	// Complete second step
	logger.Complete()

	// Output should contain both completion messages
	output := buf.String()
	if !strings.Contains(output, "Step 1 done") {
		t.Errorf("Expected output to contain 'Step 1 done', got: %s", output)
	}
	if !strings.Contains(output, "Step 2 done") {
		t.Errorf("Expected output to contain 'Step 2 done', got: %s", output)
	}
}

func TestTextLogger_Fail(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start a step
	logger.Start("Step 1", "Step 1 done")
	time.Sleep(50 * time.Millisecond)

	// Fail the step
	logger.Fail(errors.New("something went wrong"))

	// Output should contain the error message
	output := buf.String()
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("Expected output to contain 'something went wrong', got: %s", output)
	}
	if strings.Contains(output, "Step 1 done") {
		t.Errorf("Did not expect output to contain 'Step 1 done' after failure, got: %s", output)
	}
}

func TestTextLogger_FailThenStart(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start a step
	logger.Start("Step 1", "Step 1 done")
	time.Sleep(50 * time.Millisecond)

	// Fail the step
	logger.Fail(errors.New("step 1 failed"))

	// Start another step (should not try to complete the failed step)
	logger.Start("Step 2", "Step 2 done")
	time.Sleep(50 * time.Millisecond)

	// Complete second step
	logger.Complete()

	// Should contain error message and second completion
	output := buf.String()
	if !strings.Contains(output, "step 1 failed") {
		t.Errorf("Expected output to contain 'step 1 failed', got: %s", output)
	}
	if !strings.Contains(output, "Step 2 done") {
		t.Errorf("Expected output to contain 'Step 2 done', got: %s", output)
	}
	// Verify the failed step was not auto-completed
	if strings.Contains(output, "Step 1 done") {
		t.Errorf("Did not expect output to contain 'Step 1 done' after failure, got: %s", output)
	}
}

func TestTextLogger_CompleteWithoutStart(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Complete without starting (should not panic)
	logger.Complete()
}

func TestTextLogger_FailWithoutStart(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Fail without starting (should not panic)
	logger.Fail(errors.New("test error"))
}
