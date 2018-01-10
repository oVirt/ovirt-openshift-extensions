package internal

import (
	"bytes"
	"os"
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
func GetOvirtNodeName(hostname string) (string, error) {
	var err error = nil
	if hostname == "" {
		hostname, err = os.Hostname()
		if err != nil {
			return hostname, err
		}
	}
	cmd := exec.Command("kubectl", "get", "nodes", hostname, "-o", "jsonpath={.metadata.labels."+ovirtNodeNameLabel+"}")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return strings.TrimSpace(out.String()), nil
}
