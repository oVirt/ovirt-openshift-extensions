/*
Copyright 2019 oVirt-maintainers

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
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/util/file"
)

func getSystemUUIDByNodeName(nodeName string) (string, error) {
	nodes, e := getKubeNodes()
	if e != nil {
		return "", e
	}
	for _, n := range nodes {
		if n.Name == nodeName {
			return n.Status.NodeInfo.SystemUUID, nil
		}
	}
	return "", fmt.Errorf("node name %s was not found", nodeName)
}

func getKubeNodes() ([]v1.Node, error) {
	kubeconfig, err := locateKubeConfig()

	if err != nil {
		return nil, err
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

func locateKubeConfig() (string, error) {
	defaultKubeConfig := "/etc/origin/master/admin.kubeconfig"
	var err = os.ErrNotExist
	var ok bool
	if ok, err = file.FileOrSymlinkExists(defaultKubeConfig); ok {
		return defaultKubeConfig, nil
	}

	if k := os.Getenv("KUBECONFIG"); k != "" {
		if ok, err = file.FileOrSymlinkExists(k); ok {
			return k, nil
		}
	}

	if home := homeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		if ok, err = file.FileOrSymlinkExists(kubeconfig); ok {
			return kubeconfig, nil
		}
	}

	return "", err
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
