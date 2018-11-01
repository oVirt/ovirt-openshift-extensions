package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOvirtCloudProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OvirtCloudProvider Suite")
}
