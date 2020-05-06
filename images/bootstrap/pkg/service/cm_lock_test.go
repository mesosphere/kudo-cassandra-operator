package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCMUpdate_success(t *testing.T) {
	namespace = v1.NamespaceDefault
	pod = "cassandra-node-0"
	ipAddress = "10.10.10.1"
	configmap = "cassandra-topology-lock"
	bootstrapWait = "3m"
	cmLock := &v1.ConfigMapList{
		Items: []v1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cassandra-topology-lock",
					Namespace: v1.NamespaceDefault,
				},
			},
		},
	}

	fakeClient := fake.NewSimpleClientset(cmLock)
	cm := &ConfigMapLock{fakeClient}

	err := cm.UpdateCM()
	assert.Nil(t, err)
}

func TestCMUpdate_no_CM_fail(t *testing.T) {
	namespace = v1.NamespaceDefault
	pod = "cassandra-node-0"
	ipAddress = "10.10.10.1"
	configmap = "cassandra-topology-lock"
	bootstrapWait = "3m"

	fakeClient := fake.NewSimpleClientset()
	cm := &ConfigMapLock{fakeClient}

	err := cm.UpdateCM()
	assert.NotNil(t, err)
}
