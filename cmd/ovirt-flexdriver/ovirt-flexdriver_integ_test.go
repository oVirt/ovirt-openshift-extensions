/*
Copyright 2017 oVirt-maintainers

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("OVIRT_FLEXDRIVER_CONF", "../../ovirt-flexdriver.conf")
}

func TestInvocations(t *testing.T) {
	previousResult := ""
	for _, invocation := range invocationTests {
		passed := true
		if invocation.usePreviousResult != nil {
			invocation.usePreviousResult(previousResult, &invocation)
		}

		t.Logf("Test spec: %s \n", invocation.description)
		t.Logf("Test args: %s \n", invocation.args)

		r, e := App(invocation.args)

		if e != nil && invocation.exitCode == 0 {
			t.Errorf("expected a successful invocation with %s but got %s \n", invocation.args, e.Error())
			passed = false
		}
		if e == nil && invocation.exitCode > 0 {
			t.Errorf("expected a failed invocation")
			passed = false
		}
		if r == "" && e == nil {
			t.Errorf("expected some output, got none")
			passed = false
		}

		t.Logf("Test response: %v error: %v\n", r, e)

		t.Logf("Test passed: %v", passed)
		t.Logf("\n")
		previousResult = r
	}
}
