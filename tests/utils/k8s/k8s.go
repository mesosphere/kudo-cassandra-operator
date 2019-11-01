package k8s

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/api/apps/v1beta2"
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

func GetStatefulSetCount(name, namespace string) int {
	statefulSet := GetStatefulSet(name, namespace)
	if statefulSet == nil {
		log.Warningf("Found 0 replicas for statefulset %s in namespace %s .", name, namespace)
		return 0
	}
	log.Infof("Found %d  replicas of the %s in %s namespace", *statefulSet.Spec.Replicas, name, namespace)
	return int(*statefulSet.Spec.Replicas)
}

func GetStatefulSet(name, namespace string) *v1beta2.StatefulSet {
	statefulSet, err := clientset.AppsV1beta2().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Warningf("%v", err)
		return nil
	}
	return statefulSet
}

func WaitForStatefulSetReadyReplicasCount(name, namespace string, count int, timeoutSeconds time.Duration) error {
	timeout := time.After(timeoutSeconds * time.Second)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New(fmt.Sprintf("Timeout while waiting for statefulset [%s/%s] ready replicas count to be %d", namespace, name, count))
		case <-tick:
			if count == GetStatefulSetReadyReplicasCount(name, namespace) {
				return nil
			}
		}
	}
}

func GetStatefulSetReadyReplicasCount(name, namespace string) int {
	statefulSet := GetStatefulSet(name, namespace)
	if statefulSet == nil {
		log.Warningf("Found 0 replicas for statefulset %s in %s namespace.", name, namespace)
		return 0
	}
	log.Infof("Found %d/%d ready replicas of the %s in %s namespace.", statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas, name, namespace)
	return int(statefulSet.Status.ReadyReplicas)
}
