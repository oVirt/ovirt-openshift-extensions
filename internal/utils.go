package internal

import (
	"bytes"
	"os/exec"
	"strings"
)

const ovirtNodeNameLabel = "ovirtNodeNameLabel"

// GetOvirtNodeName returns the k8s ExternalID as fetched by kubectl describe node IP
// One of its usages is to bridge the gap when invoking some flexdriver command and we need the name of the
// co-responding VM
// This call expects this label to be set on the node: $ovirtNodeNameLabel
// TODO fragile approach, what happens if the node is not deployed under its `hostname`? This must be
// communicated as well in the deployment section.
func GetOvirtNodeName(hostname string) string {
	if hostname == "" {
		hostname = GetHostName()
	}
	cmd := exec.Command("kubectl", "get", "nodes", hostname, "-o", "jsonpath={.metadata.labels."+ovirtNodeNameLabel+"}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(out.String())
}

// GetHostName will return the hostname as returned by successful hostname command invocation else localhost
func GetHostName() string {
	var out bytes.Buffer
	hostname := exec.Command("hostname")
	hostname.Stdout = &out
	err := hostname.Run()
	if err != nil {
		return "localhost"
	}
	return out.String()
}
