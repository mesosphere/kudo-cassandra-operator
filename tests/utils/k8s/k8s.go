package k8s

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	cmd "github.com/mesosphere/kudo-cassandra-operator/tests/utils/cmd"
	kubectl "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
)

var (
	clientset      *kubernetes.Clientset
	kubectlOptions *kubectl.KubectlOptions
)

// Init TODO function comment.
// TODO(mpereira) return error?
func Init(_kubectlOptions *kubectl.KubectlOptions) {
	kubectlOptions = _kubectlOptions
	// TODO(mpereira) handle error.
	clientset, _ = kubectl.GetKubernetesClientFromOptions(_kubectlOptions)
}

// CreateNamespace TODO function comment.
func CreateNamespace(namespaceName string) error {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	log.Infof("Creating namespace '%s'", namespaceName)
	_, err := clientset.CoreV1().Namespaces().Create(&namespace)

	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			log.Warnf("Namespace '%s' already exists, skipping", namespaceName)
		} else {
			log.Errorf("Error creating namespace '%s': %v", namespaceName, err)
		}
		return err
	}

	log.Infof("Created namespace '%s'", namespaceName)
	return err
}

// DeleteNamespace TODO function comment.
func DeleteNamespace(namespaceName string) error {
	log.Infof("Deleting namespace '%s'", namespaceName)

	err := clientset.CoreV1().Namespaces().Delete(
		namespaceName,
		&metav1.DeleteOptions{},
	)
	if err != nil {
		log.Warnf("Error deleting namespace '%s': %s", namespaceName, err)
		return err
	}

	log.Infof("Deleted namespace '%s'", namespaceName)
	return nil
}

// GetPodContainerLogs TODO function comment.
// TODO(mpereira): use client libraries instead of shelling out.
// See: https://github.com/kubernetes/dashboard/blob/377842ddda5ce5a58e2d5397dffb14de9522ddb4/src/app/backend/resource/container/logs.go#L116
func GetPodContainerLogs(
	namespaceName string,
	podName string,
	containerName string,
) (*bytes.Buffer, error) {
	kubectlParameters := []string{
		"logs",
		podName,
		fmt.Sprintf("--namespace=%s", namespaceName),
		fmt.Sprintf("--container=%s", containerName),
	}

	_, stdout, _, err := cmd.Exec(
		kubectlOptions.KubectlPath, kubectlParameters, nil, true,
	)
	if err != nil {
		log.Errorf(
			"Error getting logs (container='%s', pod='%s', namespace='%s'): %s",
			containerName, podName, namespaceName, err,
		)
		return &bytes.Buffer{}, err
	}

	return stdout, nil
}

// ExecInPodContainer TODO function comment.
// TODO(mpereira): use client libraries instead of shelling out.
func ExecInPodContainer(
	namespaceName string,
	podName string,
	containerName string,
	command []string,
) (*bytes.Buffer, error) {
	kubectlParameters := []string{
		"exec",
		podName,
		fmt.Sprintf("--namespace=%s", namespaceName),
		fmt.Sprintf("--container=%s", containerName),
		"--",
	}
	kubectlParameters = append(kubectlParameters, command...)

	_, stdout, _, err := cmd.Exec(
		kubectlOptions.KubectlPath, kubectlParameters, nil, true,
	)
	if err != nil {
		log.Errorf(
			"Error executing '%s' (container='%s', pod='%s', namespace='%s'): %s",
			command, containerName, podName, namespaceName, err,
		)
		return &bytes.Buffer{}, err
	}

	return stdout, nil
}
