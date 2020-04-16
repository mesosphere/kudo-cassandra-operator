package cassandra

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
)

type Runner interface {
	Run(arguments ...string) (string, string, error)
}

type Executor struct {
	client    client.Client
	namespace string
}

func (c *Executor) Run(arguments ...string) (string, string, error) {
	podName := fmt.Sprintf("nodetool-%s", uuid.NewUUID())

	podTpl := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: c.namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "nodetool",
					Image:   "cassandra:3.11.5",
					Command: []string{"sleep", "3000000"},
				},
			},
		},
	}

	pod, err := kubernetes.NewPod(c.client, podTpl)
	if err != nil {
		return "", "", fmt.Errorf("failed to create new nodetool pod: %v", err)
	}
	defer func() {
		fErr := pod.Delete()
		if fErr != nil {
			fmt.Printf("Failed to delete temporary nodetool pod %s: %v", pod.Name, fErr)
		}
	}()

	for {
		if pod.Status.Phase == v1.PodRunning {
			break
		} else {
			err = pod.Update()
			if err != nil {
				return "", "", fmt.Errorf("failed to update pod: %v", err)
			}
		}
	}

	var stdOut strings.Builder
	var stdErr strings.Builder

	cmd := cmd.New("nodetool").
		WithArguments(arguments...).
		WithStdout(&stdOut).
		WithStderr(&stdErr)

	err = pod.ContainerExec("nodetool", cmd)
	if err != nil {
		err = fmt.Errorf("failed to exec in container: %v", err)
	}

	return stdOut.String(), stdErr.String(), err
}
