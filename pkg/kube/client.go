package kube

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient returns Kubernetes client to use for the worker.
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getConfig() (*rest.Config, error) {
	configPath := os.Getenv("KUBECONFIG")
	if configPath == "" {
		configPath = path.Join(os.Getenv("HOME"), ".kube/config")
	}
	if _, err := os.Stat(configPath); err == nil {
		logrus.WithField("path", configPath).Info("Using Kubernetes config based on config file")
		return clientcmd.BuildConfigFromFlags("", configPath)
	}
	logrus.Info("Using Kubernetes in-cluster config")
	return rest.InClusterConfig()
}
