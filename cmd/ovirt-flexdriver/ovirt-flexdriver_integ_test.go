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
	for _, invocation := range invocationTests {
		t.Logf("Test spec: %s \n", invocation.description)
		t.Logf("Test args: arguments %s \n", os.Args)
		r, e := App(invocation.args)
		if e != nil && invocation.exitCode == 0 {
			t.Errorf("expected a successful invocation with %s but got %s \n", invocation.args, e.Error())
		}
		if e == nil && invocation.exitCode > 0 {
			t.Errorf("expected a failed invocation with %s ", invocation.args)
		}
		if r == "" {
			t.Errorf("expected some output, got none")
		}
		t.Logf("Test says: %s\n\n", r)
	}
}
