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
	"context"
	"testing"
)

func TestWithLogger_AndFromContext(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewTextLogger(buf)

	// Add logger to context
	ctx := WithLogger(context.Background(), logger)

	// Retrieve logger from context
	retrieved := FromContext(ctx)

	// Should be the same logger
	if retrieved != logger {
		t.Error("Expected to retrieve the same logger from context")
	}
}

func TestFromContext_NoLogger(t *testing.T) {
	t.Parallel()

	// Create context without logger
	ctx := context.Background()

	// Should return NoOpLogger
	retrieved := FromContext(ctx)

	// Verify it's a noopLogger
	if _, ok := retrieved.(*noopLogger); !ok {
		t.Error("Expected NoOpLogger when context has no logger")
	}

	retrieved.Start("test", "test done")
	retrieved.Complete()
	retrieved.Fail(nil)
}

func TestFromContext_WrongType(t *testing.T) {
	t.Parallel()

	// Create context with wrong value type
	ctx := context.WithValue(context.Background(), contextKey{}, "not a logger")

	// Should return NoOpLogger
	retrieved := FromContext(ctx)

	// Verify it's a NoOpLogger
	if _, ok := retrieved.(*noopLogger); !ok {
		t.Error("Expected NoOpLogger when context has no logger")
	}

	retrieved.Start("test", "test done")
	retrieved.Complete()
	retrieved.Fail(nil)
}
