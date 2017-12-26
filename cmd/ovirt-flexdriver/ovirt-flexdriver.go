package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"pkg/Response"
)

const usage = `Usage:
	ovirt-flexdriver init
	ovirt-flexdriver getvolumename
	ovirt-flexdriver attach <json params> <nodename>
	ovirt-flexdriver detach <mount device> <nodename>
	ovirt-flexdriver waitforattach <mount device> <json params>
	ovirt-flexdriver mountdevice <mount dir> <mount device> <json params>
	ovirt-flexdriver unmountdevice <mount dir>
	ovirt-flexdriver isattached <json params> <nodename>
`

func main() {
	flag.Parse()
	args := flag.Args()
	var result Response

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

	fmt.Println(args)
	b, _ := json.Marshal(result)

	fmt.Println(string(b))

}

func initialize() Response {
	fmt.Println("initializing")
	return Response{Status: "success", Capabilities: struct{ Attach string }{"true"}}
}

func attach(jsonOpts string, nodeName string) Response {
	fmt.Printf("attaching %s %s \n", jsonOpts, nodeName)

	return Response{Status: "success", Capabilities: struct{ Attach string }{"true"}}
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
