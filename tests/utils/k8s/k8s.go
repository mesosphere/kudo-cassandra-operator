package k8s

import (
	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	kubectl "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
)

var (
	clientset      *kubernetes.Clientset
	kubectlOptions *kubectl.KubectlOptions
)

// TODO(mpereira) return error?
func Init(_kubectlOptions *kubectl.KubectlOptions) {
	kubectlOptions = _kubectlOptions
	// TODO(mpereira) handle error.
	clientset, _ = kubectl.GetKubernetesClientFromOptions(_kubectlOptions)
}

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
