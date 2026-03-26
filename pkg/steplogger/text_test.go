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
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTextLogger_CreatesSpinner(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start a step and let it run briefly
	logger.Start("Processing", "Processed")
	time.Sleep(100 * time.Millisecond)

	// Complete the step
	logger.Complete()

	// Should have output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from text logger, got empty string")
	}

	// Should contain the completion message
	if !strings.Contains(output, "Processed") {
		t.Errorf("Expected output to contain completion message 'Processed', got: %s", output)
	}
}

func TestTextLogger_UsesCompletionMessage(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start with specific messages
	logger.Start("Loading data", "Data loaded successfully")
	time.Sleep(50 * time.Millisecond)
	logger.Complete()

	output := buf.String()
	if !strings.Contains(output, "Data loaded successfully") {
		t.Errorf("Expected completion message 'Data loaded successfully', got: %s", output)
	}
}

func TestTextLogger_MultipleSteps(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// First step
	logger.Start("Step 1", "Step 1 complete")
	time.Sleep(50 * time.Millisecond)

	// Second step (auto-completes first)
	logger.Start("Step 2", "Step 2 complete")
	time.Sleep(50 * time.Millisecond)

	// Third step (auto-completes second)
	logger.Start("Step 3", "Step 3 complete")
	time.Sleep(50 * time.Millisecond)

	// Complete third step
	logger.Complete()

	output := buf.String()

	// All completion messages should be present
	if !strings.Contains(output, "Step 1 complete") {
		t.Errorf("Missing 'Step 1 complete' in output: %s", output)
	}
	if !strings.Contains(output, "Step 2 complete") {
		t.Errorf("Missing 'Step 2 complete' in output: %s", output)
	}
	if !strings.Contains(output, "Step 3 complete") {
		t.Errorf("Missing 'Step 3 complete' in output: %s", output)
	}
}

func TestTextLogger_FailShowsError(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	testErr := errors.New("database connection failed")

	logger.Start("Connecting to database", "Connected to database")
	time.Sleep(50 * time.Millisecond)
	logger.Fail(testErr)

	output := buf.String()
	if !strings.Contains(output, "database connection failed") {
		t.Errorf("Expected error message 'database connection failed', got: %s", output)
	}
}

func TestTextLogger_FailDoesNotAutoComplete(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	logger.Start("First step", "First step done")
	time.Sleep(50 * time.Millisecond)
	logger.Fail(errors.New("first step failed"))

	// Clear buffer to check second step
	buf.Reset()

	logger.Start("Second step", "Second step done")
	time.Sleep(50 * time.Millisecond)
	logger.Complete()

	output := buf.String()

	// Should NOT contain the first step's completion message
	if strings.Contains(output, "First step done") {
		t.Errorf("Failed step should not auto-complete, but output contains 'First step done': %s", output)
	}

	// Should contain the second step's completion message
	if !strings.Contains(output, "Second step done") {
		t.Errorf("Expected 'Second step done', got: %s", output)
	}
}

func TestTextLogger_DeferPattern(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	// Simulate the defer Complete() pattern
	func() {
		logger := NewTextLogger(buf)
		defer logger.Complete()

		logger.Start("Processing request", "Request processed")
		time.Sleep(50 * time.Millisecond)

		logger.Start("Saving to database", "Saved to database")
		time.Sleep(50 * time.Millisecond)
		// defer Complete() will complete the last step
	}()

	output := buf.String()

	// Both steps should be completed
	if !strings.Contains(output, "Request processed") {
		t.Errorf("Missing 'Request processed' in output: %s", output)
	}
	if !strings.Contains(output, "Saved to database") {
		t.Errorf("Missing 'Saved to database' in output: %s", output)
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

func TestTextLogger_ThreadSafety(t *testing.T) {
	t.Parallel()

	buf := &syncBuffer{}
	logger := NewTextLogger(buf)

	var wg sync.WaitGroup

	// Start multiple goroutines that use the logger
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Start(
				fmt.Sprintf("Step %d", n),
				fmt.Sprintf("Step %d complete", n),
			)
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	logger.Complete()

	// Should not panic and should have output
	output := buf.String()
	if output == "" {
		t.Error("Expected some output from concurrent operations")
	}
}

func TestTextLogger_EmptyMessages(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Start with empty messages (edge case)
	logger.Start("", "")
	time.Sleep(50 * time.Millisecond)
	logger.Complete()

	// Should not panic
}

func TestTextLogger_MultipleCompletes(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	logger.Start("Step 1", "Step 1 done")
	time.Sleep(50 * time.Millisecond)

	// Complete multiple times (should be safe)
	logger.Complete()
	logger.Complete()
	logger.Complete()

	// Should not panic
}

func TestTextLogger_MultipleFails(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	logger.Start("Step 1", "Step 1 done")
	time.Sleep(50 * time.Millisecond)

	// Fail multiple times (only first should have effect)
	logger.Fail(errors.New("first error"))
	logger.Fail(errors.New("second error"))
	logger.Fail(errors.New("third error"))

	output := buf.String()

	// Should contain first error
	if !strings.Contains(output, "first error") {
		t.Errorf("Expected 'first error', got: %s", output)
	}
	// Subsequent errors should not be recorded
	if strings.Contains(output, "second error") {
		t.Errorf("Unexpected 'second error' in output: %s", output)
	}
	if strings.Contains(output, "third error") {
		t.Errorf("Unexpected 'third error' in output: %s", output)
	}
}

func TestTextLogger_StartAfterComplete(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	logger.Start("Step 1", "Step 1 done")
	time.Sleep(50 * time.Millisecond)
	logger.Complete()

	// Start another step after completing (should work)
	logger.Start("Step 2", "Step 2 done")
	time.Sleep(50 * time.Millisecond)
	logger.Complete()

	output := buf.String()
	if !strings.Contains(output, "Step 1 done") {
		t.Errorf("Missing 'Step 1 done' in output: %s", output)
	}
	if !strings.Contains(output, "Step 2 done") {
		t.Errorf("Missing 'Step 2 done' in output: %s", output)
	}
}

func TestTextLogger_RapidStartComplete(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Rapidly start and complete steps
	for i := 0; i < 5; i++ {
		logger.Start("Quick step", "Quick step done")
		time.Sleep(5 * time.Millisecond)
	}
	logger.Complete()

	output := buf.String()
	// Should have output without panicking
	if output == "" {
		t.Error("Expected output from rapid operations")
	}
}
