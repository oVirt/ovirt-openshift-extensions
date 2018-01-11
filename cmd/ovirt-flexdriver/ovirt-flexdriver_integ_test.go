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

func TestIntegration(t *testing.T) {
	previousResult := ""
	for _, spec := range testSpecs {
		t.Run(spec.description, func(t *testing.T) {
			if spec.usePreviousResult != nil {
				spec.usePreviousResult(previousResult, &spec)
			}

			t.Logf("testSpec args: %s \n", spec.args)

			r, e := App(spec.args)

			if e != nil && spec.exitCode == 0 {
				t.Errorf("expected a successful spec with %s but got %s \n", spec.args, e.Error())
			}
			if e == nil && spec.exitCode > 0 {
				t.Errorf("expected a failed spec")
			}
			if r == "" && e == nil {
				t.Errorf("expected some output, got none")
			}

			t.Logf("testSpec response: %v error: %v\n", r, e)

			previousResult = r
		})
	}
}
