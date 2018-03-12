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
	"github.com/rgolangh/ovirt-flexdriver/internal"
	"encoding/json"
)

var attachJson = `{
	"ovirtStorageDomain": "data1",
	"ovirtVolumeName": "testDisk-100000",
	"ovirtDiskFormat": "raw",
	"kubernetes.io/fsType": "ext4",
	"kubernetes.io/readwrite": "rw",
	"capacity": "1Gi"
}`

type testSpec struct {
	description       string
	args              []string
	exitCode          int
	usePreviousResult func(result string, invocation *testSpec)
}

var testSpecs = []testSpec{
	{
		"init",
		[]string{"init"},
		0,
		nil,
	},
	{
		"init - don't fail if more args sent",
		[]string{"init", "{}"},
		0,
		nil,
	},
	{
		"attach",
		[]string{"attach"},
		1,
		nil,
	},
	{
		"attach to non existing vm",
		[]string{"attach", attachJson, "NON_EXISTING_VM"},
		1,
		nil,
	},
	{
		"attach to a vm",
		[]string{"attach", attachJson, "test_vm"},
		0,
		nil,
	},
	{
		"wait for attach",
		[]string{"waitforattach", "PREV_RESULT", "{}"},
		0,
		func(result string, invocation *testSpec) {
			r := internal.Response{}
			json.Unmarshal([]byte(result), &r)
			invocation.args[1] = r.Device
		},
	},
	{
		"get volume name",
		[]string{"getvolumename", attachJson},
		0,
		nil,
	},
	{
		"get volume name - non existing volume",
		[]string{"getvolumename", `{"ovirtVolumeName":"non-existing"}`},
		1,
		nil,
	},
}
