package kubectl

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubectlOptions struct {
	KubectlPath string
	ConfigPath  string
	Namespace   string
	Env         map[string]string
	ContextName string
}

func GetFirstNonEmptyEnvVarOrEmptyString(envVarNames []string) string {
	for _, name := range envVarNames {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}

	return ""
}

func KubeConfigPathFromHomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, ".kube", "config")
	return configPath, err
}

func GetKubeConfigPath() (string, error) {
	kubeConfigPath := GetFirstNonEmptyEnvVarOrEmptyString([]string{"KUBECONFIG"})
	if kubeConfigPath == "" {
		configPath, err := KubeConfigPathFromHomeDir()
		if err != nil {
			return "", err
		}
		kubeConfigPath = configPath
	}
	return kubeConfigPath, nil
}

func (kubectlOptions *KubectlOptions) GetConfigPath() (string, error) {
	var err error

	kubeConfigPath := kubectlOptions.ConfigPath
	if kubeConfigPath == "" {
		kubeConfigPath, err = GetKubeConfigPath()
		if err != nil {
			return "", err
		}
	}
	return kubeConfigPath, nil
}

func LoadApiClientConfig(
	configPath string, contextName string,
) (*rest.Config, error) {
	overrides := clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath},
		&overrides,
	)
	return config.ClientConfig()
}

func GetKubernetesClientFromOptions(
	options *KubectlOptions,
) (*kubernetes.Clientset, error) {
	var err error

	kubeConfigPath, err := options.GetConfigPath()
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf(
		"Configuring kubectl using config file '%s' with context '%s'",
		kubeConfigPath,
		options.ContextName,
	))

	config, err := LoadApiClientConfig(kubeConfigPath, options.ContextName)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func NewKubectlOptions(
	kubectlPath string,
	configPath string,
	namespace string,
	contextName string,
) *KubectlOptions {
	return &KubectlOptions{
		KubectlPath: kubectlPath,
		ConfigPath:  configPath,
		Namespace:   namespace,
		Env:         map[string]string{},
		ContextName: contextName,
	}
}

func BuildKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		log.Infof("Using kubeconfig at '%s'", kubeConfigPath)
		client, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Errorf("Error creating kubeconfig from %s: %v", kubeConfigPath, err)
			return nil, err
		}
		return client, err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("Error getting kubeconfig from InClusterConfig: %v", err)
		return nil, err
	}

	log.Infof("Using kubeconfig from InClusterConfig.")
	return config, err
}
