// Copyright 2026 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package podman

import "github.com/kortex-hub/kortex-cli/pkg/steplogger"

// fakeStepLogger is a fake implementation of steplogger.StepLogger that records calls for testing.
type fakeStepLogger struct {
	startCalls    []stepCall
	failCalls     []error
	completeCalls int
}

type stepCall struct {
	inProgress string
	completed  string
}

// Ensure fakeStepLogger implements steplogger.StepLogger at compile time.
var _ steplogger.StepLogger = (*fakeStepLogger)(nil)

func (f *fakeStepLogger) Start(inProgress, completed string) {
	f.startCalls = append(f.startCalls, stepCall{
		inProgress: inProgress,
		completed:  completed,
	})
}

func (f *fakeStepLogger) Fail(err error) {
	f.failCalls = append(f.failCalls, err)
}

func (f *fakeStepLogger) Complete() {
	f.completeCalls++
}
