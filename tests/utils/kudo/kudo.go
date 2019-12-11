package kudo

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	retry "github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	"github.com/kudobuilder/kudo/pkg/client/clientset/versioned"

	cmd "github.com/mesosphere/kudo-cassandra-operator/tests/utils/cmd"
	k8s "github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	kubectl "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
)

var (
	kubectlOptions *kubectl.KubectlOptions
	kudo           *versioned.Clientset
)

// Init TODO function comment.
// TODO(mpereira) return err?
func Init(_kubectlOptions *kubectl.KubectlOptions) {
	kubectlOptions = _kubectlOptions
	// TODO(mpereira) handle err.
	kubeconfig, _ := kubectl.BuildKubeConfig(kubectlOptions.ConfigPath)
	// TODO(mpereira) handle err.
	kudo, _ = versioned.NewForConfig(kubeconfig)
}

// GetInstance TODO function comment.
func GetInstance(
	namespaceName string, instanceName string,
) (*v1beta1.Instance, error) {
	instances := kudo.KudoV1beta1().Instances(namespaceName)
	instance, err := instances.Get(instanceName, metav1.GetOptions{})

	if err != nil {
		log.Errorf(
			"Error getting KUDO instance (namespace='%s', name='%s'): %v",
			namespaceName,
			instanceName,
			err,
		)
		return nil, err
	}

	if instance == nil {
		log.Warnf(
			"No KUDO instance found (namespace='%s', name='%s')",
			namespaceName,
			instanceName,
		)
		return nil, err
	}

	return instance, nil
}

// GetInstanceAggregatedStatus TODO function comment.
func GetInstanceAggregatedStatus(
	namespaceName string, instanceName string,
) (*v1beta1.ExecutionStatus, error) {
	instance, err := GetInstance(namespaceName, instanceName)

	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, err
	}

	return &instance.Status.AggregatedStatus.Status, err
}

// WaitForOperatorDeployStatus TODO function comment.
func WaitForOperatorDeployStatus(
	expectedStatus v1beta1.ExecutionStatus,
	namespaceName string,
	instanceName string,
	retryDelay time.Duration,
	retryAttempts uint,
) error {
	return retry.Do(
		func() error {
			status, err := GetInstanceAggregatedStatus(namespaceName, instanceName)

			if err != nil {
				log.Errorf(
					"Error attempting to get operator (instance='%s', namespace='%s') "+
						"deploy status: %s",
					instanceName,
					namespaceName,
					err,
				)
				return errors.New("")
			}

			if status == nil {
				log.Warnf(
					"Waiting for operator (instance='%s', namespace='%s') deploy status "+
						"to be '%s', is not available",
					instanceName,
					namespaceName,
					expectedStatus,
				)
				return errors.New("")
			}

			if expectedStatus != *status {
				log.Infof(
					"Waiting for operator (instance='%s', namespace='%s') deploy status "+
						"to be '%s', is '%s'",
					instanceName,
					namespaceName,
					expectedStatus,
					*status,
				)
				return errors.New("")
			}

			log.Infof(
				"Operator (instance='%s', namespace='%s') deploy status is '%s'",
				instanceName,
				namespaceName,
				*status,
			)
			return nil
		},
		retry.DelayType(retry.FixedDelay),
		retry.Delay(retryDelay),
		retry.Attempts(retryAttempts),
	)
}

// WaitForOperatorDeployInProgress TODO function comment.
func WaitForOperatorDeployInProgress(
	namespaceName string, instanceName string,
) error {
	// 30 seconds.
	retryDelay := time.Second * 3
	var retryAttempts uint = 10

	return WaitForOperatorDeployStatus(
		v1beta1.ExecutionInProgress,
		namespaceName,
		instanceName,
		retryDelay,
		retryAttempts,
	)
}

// WaitForOperatorDeployComplete TODO function comment.
func WaitForOperatorDeployComplete(
	namespaceName string, instanceName string,
) error {
	// 5 minutes.
	retryDelay := time.Second * 10
	var retryAttempts uint = 30

	return WaitForOperatorDeployStatus(
		v1beta1.ExecutionComplete,
		namespaceName,
		instanceName,
		retryDelay,
		retryAttempts,
	)
}

// UpdateInstanceParameters TODO function comment.
func UpdateInstanceParameters(
	namespaceName string, instanceName string, parameters map[string]string,
) error {
	log.Infof(
		"Updating instance (instance='%s', namespace='%s')",
		namespaceName, instanceName,
	)

	instances := kudo.KudoV1beta1().Instances(namespaceName)
	instance, err := instances.Get(instanceName, metav1.GetOptions{})
	if err != nil {
		log.Errorf(
			"Error getting instance (instance='%s', namespace='%s'): %v",
			namespaceName, instanceName, err,
		)
		return err
	}

	if instance == nil {
		log.Warnf(
			"Instance not found (instance='%s', namespace='%s')",
			namespaceName, instanceName,
		)
		return errors.New("Instance not found")
	}

	newParameters := make(map[string]string)
	for k, v := range instance.Spec.Parameters {
		newParameters[k] = v
	}

	for k, v := range parameters {
		log.Infof(
			"Will update '%s' from '%s' to '%s'", k, instance.Spec.Parameters[k], v,
		)
		newParameters[k] = v
	}
	instance.Spec.Parameters = newParameters

	_, err = instances.Update(instance)
	if err != nil {
		log.Errorf(
			"Error updating instance (instance='%s', namespace='%s'): %v",
			namespaceName, instanceName, err,
		)
		return err
	}

	err = WaitForOperatorDeployInProgress(namespaceName, instanceName)
	if err != nil {
		log.Errorf("Error waiting for operator deploy to be in-progress: %s", err)
		return err
	}

	err = WaitForOperatorDeployComplete(namespaceName, instanceName)
	if err != nil {
		log.Errorf("Error waiting for operator deploy to complete: %s", err)
		return err
	}

	log.Infof(
		"Updated instance (instance='%s', namespace='%s')",
		instanceName, namespaceName,
	)

	return nil
}

func installOrUpgradeOperator(
	installOrUpgrade string,
	operatorNameOrDirectory string,
	namespaceName string,
	instanceName string,
	parameters []string,
) error {
	if installOrUpgrade != "install" && installOrUpgrade != "upgrade" {
		fmt.Errorf(
			"Expected 'installOrUpgrade' to be 'install' or 'upgrade', was: '%s'",
			installOrUpgrade,
		)
	}

	log.Infof(
		"%sing '%s' operator (instance='%s', namespace='%s')",
		strings.Title(installOrUpgrade),
		operatorNameOrDirectory,
		instanceName,
		namespaceName,
	)

	kubectlParameters := []string{
		"kudo",
		installOrUpgrade,
		operatorNameOrDirectory,
		fmt.Sprintf("--namespace=%s", namespaceName),
		fmt.Sprintf("--instance=%s", instanceName),
	}

	for _, parameter := range parameters {
		kubectlParameters = append(
			kubectlParameters, "--parameter", string(parameter),
		)
	}

	_, _, _, err := cmd.Exec(
		kubectlOptions.KubectlPath, kubectlParameters, nil, false,
	)
	if err != nil {
		log.Errorf(
			"Error trying to %s '%s' operator: %s",
			installOrUpgrade, operatorNameOrDirectory, err,
		)
		return err
	}

	log.Infof(
		"Started '%s' operator %s (instance='%s', namespace='%s')",
		operatorNameOrDirectory, installOrUpgrade, instanceName, namespaceName,
	)

	err = WaitForOperatorDeployInProgress(namespaceName, instanceName)
	if err != nil {
		log.Errorf("Error waiting for operator deploy to be in-progress: %s", err)
		return err
	}

	err = WaitForOperatorDeployComplete(namespaceName, instanceName)
	if err != nil {
		log.Errorf(
			"Error waiting for '%s' operator deploy to complete: %s",
			operatorNameOrDirectory, err,
		)
		return err
	}

	return nil
}

// InstallOperator TODO function comment.
func InstallOperator(
	operatorNameOrDirectory string,
	namespaceName string,
	instanceName string,
	parameters []string,
) error {
	return installOrUpgradeOperator(
		"install", operatorNameOrDirectory, namespaceName, instanceName, parameters,
	)
}

// UpgradeOperator TODO function comment.
func UpgradeOperator(
	operatorNameOrDirectory string,
	namespaceName string,
	instanceName string,
	parameters []string,
) error {
	return installOrUpgradeOperator(
		"upgrade", operatorNameOrDirectory, namespaceName, instanceName, parameters,
	)
}

// UninstallOperator TODO function comment.
func UninstallOperator(
	operatorName string, namespaceName string, instanceName string,
) error {
	uninstallScript := "../../scripts/uninstall_operator.sh"
	uninstallScriptParameters := []string{
		"--operator", operatorName,
		"--instance", instanceName,
		"--namespace", namespaceName,
	}

	log.Infof(
		"Uninstalling '%s' (instance='%s', namespace='%s')",
		operatorName, instanceName, namespaceName,
	)

	_, _, _, err := cmd.Exec(
		uninstallScript,
		uninstallScriptParameters,
		[]string{fmt.Sprintf("KUBECTL_PATH=%s", kubectlOptions.KubectlPath)},
		false,
	)
	if err != nil {
		log.Errorf(
			"Error trying to uninstall '%s' (instance='%s', namespace='%s'): %s",
			operatorName, instanceName, namespaceName, err,
		)
		return err
	}

	log.Infof(
		"Successfully uninstalled '%s' (instance='%s', namespace='%s')",
		operatorName, instanceName, namespaceName,
	)

	return nil
}

// GetPodContainerLogs TODO function comment.
func GetPodContainerLogs(
	namespaceName string,
	instanceName string,
	podName string,
	podInstance int,
	containerName string,
) (*bytes.Buffer, error) {
	return k8s.GetPodContainerLogs(
		namespaceName,
		fmt.Sprintf("%s-%s-%d", instanceName, podName, podInstance),
		containerName,
	)
}

// ExecInPodContainer TODO function comment.
func ExecInPodContainer(
	namespaceName string,
	instanceName string,
	podName string,
	podInstance int,
	containerName string,
	command []string,
) (*bytes.Buffer, error) {
	return k8s.ExecInPodContainer(
		namespaceName,
		fmt.Sprintf("%s-%s-%d", instanceName, podName, podInstance),
		containerName,
		command,
	)
}
