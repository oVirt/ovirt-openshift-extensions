package internal

import "testing"

func TestGetExternalNodeName(t *testing.T) {
	nodeName := GetOvirtNodeName()
	t.Logf("node name: %s", nodeName)
	if nodeName == "" {
		t.Error("Expected to get some node name")
	}
}
