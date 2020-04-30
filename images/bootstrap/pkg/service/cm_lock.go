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

func (c *ConfigMapLock) AcquireLock() bool {
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("bootstrap: configmap cassandra-topology configmap %s cannot be found...", configmap)
		return false
	}
	if cfg.Annotations == nil {
		cfg.Annotations = make(map[string]string)
	}
	cfg.Name = configmap
	if cfg.Annotations[annotationLock] == "" {
		cfg.Annotations[annotationLock] = pod
		log.Infof("bootstrap: acquiring annotationLock of configmap %s/%s", namespace, configmap)
		_, err = c.CoreV1().ConfigMaps(namespace).Update(cfg)
		return err == nil
	}
	return false
}

func (c *ConfigMapLock) HasLock() bool {
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("bootstrap: configmap cassandra-topology configmap %s cannot be found...", configmap)
		return false
	}
	if cfg.Annotations == nil {
		return false
	}
	return cfg.Annotations[annotationLock] == pod
}

func (c *ConfigMapLock) ReleaseLock() bool {
	if c.HasLock() {
		cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
		if errors.IsNotFound(err) {
			log.Warnf("bootstrap: configmap cassandra-topology configmap %s cannot be found...", configmap)
		}
		if cfg.Annotations == nil {
			cfg.Annotations = make(map[string]string)
		}
		cfg.Annotations[annotationLock] = ""
		_, err = c.CoreV1().ConfigMaps(namespace).Update(cfg)
		return err == nil
	}
	return false
}

func (c *ConfigMapLock) UpdateCM() (bool, error) {
	c.AcquireLock()
	defer c.ReleaseLock()
	if !c.HasLock() {
		return false, fmt.Errorf("pod %s doesn't have the annotationLock", pod)
	}
	cfg, err := c.GetConfigMap(namespace, configmap)
	if errors.IsNotFound(err) {
		log.Warnf("bootstrap: configmap cassandra-topology configmap %s cannot be found...", configmap)
	}
	_, err = c.UpdateConfigMap(namespace, cfg)
	if err != nil {
		return false, err
	}
	return true, err
}

func (c *ConfigMapLock) GetConfigMap(ns string, name string) (*v1.ConfigMap, error) {
	return c.CoreV1().ConfigMaps(ns).Get(name, meta_v1.GetOptions{})
}

func (c *ConfigMapLock) UpdateConfigMap(ns string, cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[pod] = ipAddress
	cm.Data["last-updated-by"] = pod
	log.Infof("bootstrap: Updating configmap %s/%s with IP [%s] for pod [%s]\n", ns, cm.GetName(), ipAddress, pod)
	return c.CoreV1().ConfigMaps(ns).Update(cm)
}
