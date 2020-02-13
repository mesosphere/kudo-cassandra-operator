package curl

import (
	"fmt"
	"strings"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func RunCommand(client client.Client, namespace string, arguments ...string) (string, string, error) {
	curlPodName := fmt.Sprintf("curl-%s", uuid.NewUUID())

	podTpl := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      curlPodName,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "curl",
					Image:   "curlimages/curl",
					Command: []string{"sleep", "300"},
				},
			},
		},
	}

	pod, err := kubernetes.NewPod(client, podTpl)
	if err != nil {
		return "", "", fmt.Errorf("failed to create new curl pod: %v", err)
	}
	defer func() {
		_ = pod.Delete()
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

	var stdout strings.Builder
	var stderr strings.Builder

	cmd := cmd.New("curl").
		WithArguments(arguments...).
		WithStdout(&stdout).
		WithStderr(&stderr)

	err = pod.ContainerExec("curl", cmd)
	if err != nil {
		return "", "", fmt.Errorf("failed to exec in container: %v", err)
	}

	return stdout.String(), stderr.String(), nil
}
