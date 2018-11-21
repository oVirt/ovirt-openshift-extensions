/*
Copyright 2018 oVirt-maintainers

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
	"flag"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/ovirt/ovirt-openshift-extensions/internal"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var (

	// Name of the provisioner.
	// The provisioner will only provision volumes for claims that
	// request a StorageClass with a provisioner field set equal to this name.
	provisioner = "external/ovirt"
	master      = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig  = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	glog.Infof("Provisioner %s specified", provisioner)

	clientSet, serverVersion := getClientSet()
	ovirtApi, err := newOvirt()
	if err != nil {
		glog.Fatalf("Failed to initialize ovirt client: %v", err)
	}
	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	ovirtProvisioner := NewOvirtProvisioner(ovirtApi)

	// Start the provision controller which will dynamically provision NFS PVs
	pc := controller.NewProvisionController(
		clientSet,
		provisioner,
		ovirtProvisioner,
		serverVersion.GitVersion,
	)

	pc.Run(wait.NeverStop)
}
func getClientSet() (kubernetes.Interface, version.Info) {
	// Create the client according to whether we are running in or out-of-cluster
	var config *rest.Config
	var err error
	if *master != "" || *kubeconfig != "" {
		glog.Infof("Either master or kubeconfig specified. building kube config from that..")
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		glog.Infof("Building kube configs for running in cluster...")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %v", err)
	}
	return clientset, *serverVersion
}

func newOvirt() (internal.OvirtApi, error) {
	var conf string
	value, exist := os.LookupEnv("OVIRT_API_CONF")
	if exist {
		conf = value
	} else {
		conf = "/etc/ovirt/ovirt-api.conf"
	}
	file, e := os.Open(conf)
	if e != nil {
		return nil, e
	}
	ovirt, err := internal.NewOvirt(file)
	if err != nil {
		return nil, err
	}
	err = ovirt.Authenticate()
	if err != nil {
		return nil, err
	}
	return ovirt, nil
}
