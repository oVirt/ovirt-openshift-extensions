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
func GetOvirtNodeName() string {
	var outa bytes.Buffer
	hostname := exec.Command("hostname")
	hostname.Stdout = &outa
	err := hostname.Run()
	if err != nil {
		return err.Error()
	}
	h := outa.String()

	cmd := exec.Command("kubectl", "get", "nodes", h, "-o", "jsonpath={.metadata.labels."+ovirtNodeNameLabel+"}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(out.String())
}
