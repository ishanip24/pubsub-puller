package controllers

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	coreClient "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restClient "k8s.io/client-go/rest"
	cmdClient "k8s.io/client-go/tools/clientcmd"
)

// NewKubeClient will provide a new k8s client interface
// resolves where it is running whether inside the kubernetes cluster or outside
// While running outside of the cluster, tries to make use of the kubeconfig file
// While running inside the cluster resolved via pod environment uses the in-cluster config
func NewKubeClient() (*coreClient.Clientset, error) {
	config, err := restClient.InClusterConfig()
	if err != nil {
		kubeconfig := cmdClient.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = cmdClient.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return coreClient.NewForConfig(config)
}

// GetVersion returns the version of the kubernetes cluster that is running
func GetVersion(logger *zap.Logger, clientSet *coreClient.Clientset) (string, error) {
	version, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("Version of running k8s %v", version))

	return version.String(), nil
}

// GetNamespace will return the current namespace for the running program
// Checks for the user passed ENV variable POD_NAMESPACE if not available
// pulls the namespace from pod, if not returns ""
func GetNamespace() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
		return "", err
	}
	return "", nil
}
