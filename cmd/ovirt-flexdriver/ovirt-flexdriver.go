//Copyright 2017 oVirt-maintainers
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ovirt/ovirt-flexdriver/internal"
	"gopkg.in/gcfg.v1"
	"os"
)

const usage = `Usage:
	ovirt-flexdriver init
	ovirt-flexdriver attach <json params> <nodename>
	ovirt-flexdriver detach <mount device> <nodename>
	ovirt-flexdriver waitforattach <mount device> <json params>
	ovirt-flexdriver mountdevice <mount dir> <mount device> <json params>
	ovirt-flexdriver unmountdevice <mount dir>
	ovirt-flexdriver isattached <json params> <nodename>
`

var driverConfig = "ovirt-flexdriver.conf"

func main() {
	flag.Parse()
	args := flag.Args()

	var result internal.Response

	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		result = initialize()
	case "attach":
		attach(args[1], args[2])
	case "detach":
		detach(args[1], args[2])
	default:
	}

	b, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed json marshalling the result %s", err)
	}
	fmt.Println(string(b))
}

func initialize() internal.Response {
	value, exist := os.LookupEnv("OVIRT_FLEXDRIVER_CONF")
	if exist {
		driverConfig = value
	}

	api := internal.Api{}
	err := gcfg.ReadFileInto(&api, driverConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading the configuration file %s", err)
		os.Exit(1)
	}

	err = api.Authenticate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

	return internal.Response{Status: "success", Capabilities: struct{ Attach string }{"true"}}
}

func attach(jsonOpts string, nodeName string) internal.Response {
	fmt.Printf("attaching %s %s \n", jsonOpts, nodeName)

	return internal.Response{Status: "success", Capabilities: struct{ Attach string }{"true"}}
}
func isattached(jsonOpts string, nodeName string) {
	fmt.Printf("isattached %s %s \n", jsonOpts, nodeName)
}

func detach(mountDevice string, nodeName string) {
	fmt.Printf("detaching %s %s \n", mountDevice, nodeName)
}

func waitForAttach(mountDevice string, nodeName string) {
	fmt.Printf("waitForAttach %s %s \n", mountDevice, nodeName)
}

func mountDevice(mountDir string, mountDevice string, jsonOpts string) {
	fmt.Printf("mountDevicee %s \n", mountDevice)
}

func unmountDevice(mountDevice string) {
	fmt.Printf("mountDevicee %s \n", mountDevice)
}
