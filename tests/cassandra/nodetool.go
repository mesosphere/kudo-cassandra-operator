package cassandra

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
)

// Runner interface
type Runner interface {
	Run(arguments ...string) (string, string, error)
}

// Executor struct
type Executor struct {
	client     client.Client
	namespace  string
	instance   string
	sslEnabled bool
}

// NewNodeTool Create new nodetool
func NewNodeTool(client client.Client, namespace, instance string, sslEnabled bool) Runner {
	return &Executor{
		client:     client,
		namespace:  namespace,
		instance:   instance,
		sslEnabled: sslEnabled,
	}
}

// Run Runs nodetool inside a new pod
func (c *Executor) Run(arguments ...string) (string, string, error) {
	podName := fmt.Sprintf("nodetool-%s", uuid.NewUUID())
	defaultMode := int32(0755)
	id := int64(999)
	runAsNonRoot := true
	args := "sleep 1000000;"
	volumeMounts := []v1.VolumeMount{}
	volumes := []v1.Volume{}

	if c.sslEnabled {
		args = "/etc/tls/bin/generate-tls-artifacts.sh;" +
			"/etc/cassandra/generate-nodetool-ssl-properties.sh;" +
			"cp /etc/cassandra/nodetool-ssl.properties /home/cassandra/.cassandra/nodetool-ssl.properties;" +
			args
		volumeMounts = []v1.VolumeMount{
			{
				Name:      "etc-cassandra",
				MountPath: "/etc/cassandra",
			},
			{
				Name:      "cassandra-tls",
				MountPath: "/etc/tls/certs",
			},
			{
				Name:      "generate-tls-artifacts",
				MountPath: "/etc/tls/bin",
			},
			{
				Name:      "generate-nodetool-ssl-properties",
				MountPath: "/etc/cassandra/generate-nodetool-ssl-properties.sh",
				SubPath:   "generate-nodetool-ssl-properties.sh",
			},
			{
				Name:      "dot-cassandra",
				MountPath: "/home/cassandra/.cassandra/",
			},
			{
				Name:      "truststore-credentials",
				MountPath: "/etc/cassandra/truststore",
				ReadOnly:  true,
			},
		}
		volumes = []v1.Volume{
			{
				Name: "etc-cassandra",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "cassandra-tls",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "cassandra-tls",
					},
				},
			},
			{
				Name: "generate-tls-artifacts",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: fmt.Sprintf("%s-generate-tls-artifacts-sh", c.instance),
						},
						DefaultMode: &defaultMode,
					},
				},
			},
			{
				Name: "generate-nodetool-ssl-properties",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: fmt.Sprintf("%s-generate-nodetool-ssl-properties", c.instance),
						},
						DefaultMode: &defaultMode,
					},
				},
			},
			{
				Name: "dot-cassandra",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "truststore-credentials",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: fmt.Sprintf("%s-tls-store-credentials", c.instance),
					},
				},
			},
		}
	}

	podTpl := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: c.namespace,
		},
		Spec: v1.PodSpec{
			SecurityContext: &v1.PodSecurityContext{
				RunAsNonRoot: &runAsNonRoot,
				RunAsUser:    &id,
				RunAsGroup:   &id,
				FSGroup:      &id,
			},
			Containers: []v1.Container{
				{
					Name:  "nodetool",
					Image: "cassandra:3.11.5",
					Command: []string{
						"bash",
						"-c",
					},
					Args: []string{
						args,
					},
					VolumeMounts: volumeMounts,
				},
			},
			Volumes: volumes,
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

	// The generate-tls-artifacts.sh takes time to execute in the container.
	// Waiting for it to complete before exec-ing in the container
	fmt.Printf("Waiting 10 secs for pod %s to initialize", pod.Name)
	time.Sleep(10 * time.Second)

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
