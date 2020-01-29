package kudo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	"github.com/kudobuilder/kudo/pkg/client/clientset/versioned"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/cmd"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
)

var (
	kubectlOptions       *kubectl.KubectlOptions
	kudo                 *versioned.Clientset
	operatorYamlFilePath = "../../../operator/operator.yaml"
	// https://regex101.com/r/Ly7O1x/3/
	semVerRegexp = `(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
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

// OverrideOperatorVersion rewrites `operatorVersion` in operator.yaml. This is
// currently necessary in integration tests that run upgrades to make it
// possible to upgrade to a version based on the local filesystem even when the
// `operatorVersion` is the same as the one being upgraded from.
//
// Returns the operator version parsed before overriding and the operator
// version written.
func OverrideOperatorVersion(
	desiredOperatorVersion string,
) (string, string, error) {
	operatorYamlBytes, err := ioutil.ReadFile(operatorYamlFilePath)
	if err != nil {
		log.Errorf("Failed to read '%s': %v", operatorYamlFilePath, err)
		return "", "", err
	}

	scanner := bufio.NewScanner(bytes.NewReader(operatorYamlBytes))

	operatorVersionLineRegexp := fmt.Sprintf(
		`^operatorVersion:\s*"(%s)"$`, semVerRegexp,
	)
	operatorVersionLineLinePattern := regexp.MustCompile(operatorVersionLineRegexp)

	var operatorVersion string
	for scanner.Scan() {
		text := scanner.Text()
		match := operatorVersionLineLinePattern.FindStringSubmatch(text)
		if len(match) > 0 {
			operatorVersion = match[1]
		}
	}

	if operatorVersion == "" {
		errorMessage := fmt.Sprintf(
			"Failed to parse operatorVersion from '%s'", operatorYamlFilePath,
		)
		log.Error(errorMessage)
		return "", "", errors.New(errorMessage)
	}

	updatedOperatorYamlBytes := bytes.Replace(
		operatorYamlBytes,
		[]byte(operatorVersion),
		[]byte(desiredOperatorVersion),
		-1,
	)

	err = ioutil.WriteFile(operatorYamlFilePath, updatedOperatorYamlBytes, 0644)

	if err != nil {
		log.Errorf("Failed to write updated '%s': %v", operatorYamlFilePath, err)
		return "", "", err
	}

	log.Infof(
		"Overrode operatorVersion from '%s' to '%s'",
		operatorVersion,
		desiredOperatorVersion,
	)

	return operatorVersion, desiredOperatorVersion, nil
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
) (*v1beta1.ExecutionStatus, int64, error) {
	instance, err := GetInstance(namespaceName, instanceName)

	if err != nil {
		return nil, -1, err
	}

	if instance == nil {
		return nil, -1, err
	}

	for _, plan := range instance.Status.PlanStatus {
		for _, phase := range plan.Phases {
			for _, step := range phase.Steps {
				if step.Status == v1beta1.ErrorStatus {
					log.Warnf("Plan %s, Phase %s, Step %s is in ErrorStatus: %s", plan.Name, phase.Name, step.Name, step.Message)
				}
			}
		}
	}

	planUID := instance.Generation

	return &instance.Status.AggregatedStatus.Status, planUID, err
}

// WaitForOperatorDeployStatus TODO function comment.
func WaitForOperatorDeployStatus(
	expectedStatus v1beta1.ExecutionStatus,
	namespaceName string,
	instanceName string,
	retryDelay time.Duration,
	retryAttempts uint,
	oldGeneration int64,
) error {
	return retry.Do(
		func() error {
			status, generation, err := GetInstanceAggregatedStatus(namespaceName, instanceName)

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

			if generation == oldGeneration {
				log.Warnf("Old Instance Generation '%d' equals current generation, no new plan was triggered", generation)
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
		-1,
	)
}

// WaitForOperatorDeployComplete TODO function comment.
func WaitForOperatorDeployComplete(
	namespaceName string, instanceName string, oldGeneration int64,
) error {
	// 5 minutes.
	retryDelay := time.Second * 15
	var retryAttempts uint = 30

	return WaitForOperatorDeployStatus(
		v1beta1.ExecutionComplete,
		namespaceName,
		instanceName,
		retryDelay,
		retryAttempts,
		oldGeneration,
	)
}

// UpdateInstanceParameters TODO function comment.
func UpdateInstanceParameters(
	namespaceName string, instanceName string, parameters map[string]string,
	waitForDeploy bool,
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

	// Save old Instance generation so we can verify that a new plan was executed
	oldGeneration := instance.Generation

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

	if waitForDeploy {
		err = WaitForOperatorDeployInProgress(namespaceName, instanceName)
		if err != nil {
			log.Errorf("Error waiting for operator deploy to be in-progress: %s", err)
			return err
		}
	}

	err = WaitForOperatorDeployComplete(namespaceName, instanceName, oldGeneration)
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
	waitForDeploy bool,
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

	if waitForDeploy {
		err = WaitForOperatorDeployInProgress(namespaceName, instanceName)
		if err != nil {
			log.Errorf("Error waiting for operator deploy to be in-progress: %s", err)
			return err
		}
	}

	err = WaitForOperatorDeployComplete(namespaceName, instanceName, -1)
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
	waitForDeploy bool,
) error {
	return installOrUpgradeOperator(
		"install", operatorNameOrDirectory, namespaceName, instanceName, parameters, waitForDeploy,
	)
}

// UpgradeOperator TODO function comment.
func UpgradeOperator(
	operatorNameOrDirectory string,
	namespaceName string,
	instanceName string,
	parameters []string,
	waitForDeploy bool,
) error {
	return installOrUpgradeOperator(
		"upgrade", operatorNameOrDirectory, namespaceName, instanceName, parameters, waitForDeploy,
	)
}

// UninstallOperator TODO function comment.
func UninstallOperator(
	operatorName string, namespaceName string, instanceName string,
) error {
	uninstallScript := "../../../scripts/uninstall_operator.sh"
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
