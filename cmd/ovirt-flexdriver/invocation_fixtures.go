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

var attachJson = `{
	"ovirtStorageDomain": "data",
	"ovirtVolumeName": "testDisk-100000",
	"ovirtDiskFormat": "raw",
	"kubernetes.io/fsType": "ext4",
	"kubernetes.io/readwrite": "rw"
}`

var invocationTests = []struct {
	description string
	args        []string
	exitCode    int
}{
	{
		"init",
		[]string{"init"},
		0,
	},
	{
		"init - don't fail if more args sent",
		[]string{"init", "{}"},
		0,
	},
	{
		"attach",
		[]string{"attach"},
		1,
	},
	{
		"attach to non existing vm",
		[]string{"attach", attachJson, "NON_EXISTING_VM"},
		1,
	},
	{
		"attach to a vm",
		[]string{"attach", attachJson, "test_vm"},
		0,
	},
}
