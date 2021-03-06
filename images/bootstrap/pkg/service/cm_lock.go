package service

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMapLock struct {
	kubernetes.Interface
}

func (c *ConfigMapLock) AcquireLock() (*v1.ConfigMap, error) {
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmapName, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("bootstrap: cassandra-topology configmap %s cannot be found...", configmapName)
		return nil, err
	}
	if cfg.Annotations == nil {
		cfg.Annotations = make(map[string]string)
	}
	cfg.Name = configmapName
	if cfg.Annotations[ANNOTATION_LOCK] == "" || cfg.Annotations[ANNOTATION_LOCK] == podName {
		cfg.Annotations[ANNOTATION_LOCK] = podName
		log.Infof("bootstrap: acquiring annotationLock of configmap %s/%s", namespace, configmapName)
		return c.CoreV1().ConfigMaps(namespace).Update(cfg)
	}
	return nil, fmt.Errorf("cannot acquire lock for %s. pod %s has the lock", configmapName, cfg.Annotations[ANNOTATION_LOCK])
}

func (c *ConfigMapLock) HasLock() (*v1.ConfigMap, error) {
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmapName, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("bootstrap: configmap cassandra-topology configmap %s cannot be found...", configmapName)
		return nil, err
	}
	if cfg.Annotations == nil {
		return nil, fmt.Errorf("configmap %s has no annotations", configmapName)
	}
	if cfg.Annotations[ANNOTATION_LOCK] == podName {
		return cfg, nil
	}
	return nil, fmt.Errorf("%s doesn't have the lock in %s", podName, configmapName)
}

func (c *ConfigMapLock) ReleaseLock() bool {
	cm, err := c.HasLock()
	if err != nil {
		return true
	}
	cm.Annotations[ANNOTATION_LOCK] = ""
	_, err = c.CoreV1().ConfigMaps(namespace).Update(cm)
	return err == nil
}

func (c *ConfigMapLock) UpdateCM() error {
	cm, err := c.AcquireLock()
	defer c.ReleaseLock()
	if err != nil {
		return err
	}
	_, err = c.UpdateConfigMap(namespace, cm)
	return err
}

func (c *ConfigMapLock) GetConfigMap(ns string, name string) (*v1.ConfigMap, error) {
	return c.CoreV1().ConfigMaps(ns).Get(name, meta_v1.GetOptions{})
}

func (c *ConfigMapLock) UpdateConfigMap(ns string, cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[podName] = podIpAddress
	cm.Data["last-updated-by"] = podName
	log.Infof("bootstrap: Updating configmap %s/%s with IP [%s] for pod [%s]\n", ns, cm.GetName(), podIpAddress, podName)
	return c.CoreV1().ConfigMaps(ns).Update(cm)
}
