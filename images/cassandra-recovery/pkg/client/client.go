package client

import (
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct{}

func (c *Client) buildKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		client, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Errorf("error creating kubernetes client from %s: %v", kubeconfig, err)
			return nil, err
		}
		return client, err
	}
	log.Infof("kubeconfig file: using InClusterConfig.")
	return rest.InClusterConfig()
}

func (c *Client) getKubernetesClient(kubeconfig *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Fatalf("error creating kubernetes client: %v", err)
		return nil, err
	}
	log.Infof("kubernetes client configured.")
	return clientSet, nil
}

func GetKubernetesClient() (*rest.Config, error) {
	c := Client{}
	kubeConfigPath := os.Getenv("KUBECONFIG")
	return c.buildKubeConfig(kubeConfigPath)
}
