package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	namespace   string
	pod         string
	configmap   = "cassandra-instance-topology-lock"
	ipAddress   string
	lock        = "cassandra.kudo.dev/lock"
	replaceFile = "/var/lib/cassandra/replace.ip"
)

type CassandraService struct {
	*kubernetes.Clientset
}

func (c *CassandraService) SetReplaceIP() (bool, error) {
	c.AquireLock()
	defer c.ReleaseLock()
	if !c.HasLock() {
		return false, fmt.Errorf("pod %s doesn't have the lock", pod)
	}
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("configmap cassandra-topology configmap %s cannot be found...", configmap)
	}
	if cfg.Data == nil {
		cfg.Data = make(map[string]string)
	}

	nodeIp := cfg.Data[pod]
	if nodeIp == "" {
		log.Infoln("Initializing the node values to cassandra topology")
		cfg.Data[pod] = ipAddress
		_, err = c.CoreV1().ConfigMaps(namespace).Update(cfg)
		return false, err
	} else if nodeIp != ipAddress {
		// new internal ip address
		// check if data is present
		if dataExists() {
			return true, nil
		}
		return true, c.WriteReplaceIp(nodeIp)
	}
	return false, nil
}

func dataExists() bool {
	_, err := os.Stat("/var/lib/cassandra/data/system")
	// create file if not exists
	if os.IsNotExist(err) {
		return false
	}
	if err == nil {
		return true
	} else {
		log.Errorf("Error when checking for /var/lib/cassandra/data/system %v", err)
		return false
	}

}

func (c *CassandraService) AquireLock() bool {
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("configmap cassandra-topology configmap %s cannot be found...", configmap)
	}
	if cfg.Annotations == nil {
		cfg.Annotations = make(map[string]string)
	}
	cfg.Name = configmap
	if cfg.Annotations[lock] == "" {
		cfg.Annotations[lock] = pod
		log.Infof("configmap: %+v", cfg)
		_, err = c.CoreV1().ConfigMaps(namespace).Update(cfg)
		if err != nil {
			log.Errorf("error during cm update: %v", err)
			return false
		}
		return true
	}
	return false
}

func (c *CassandraService) HasLock() bool {
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("configmap cassandra-topology configmap %s cannot be found...", configmap)
	}
	return cfg.Annotations[lock] == pod
}

func (c *CassandraService) ReleaseLock() bool {
	if c.HasLock() {
		cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
		if errors.IsNotFound(err) {
			log.Warnf("configmap cassandra-topology configmap %s cannot be found...", configmap)
		}
		cfg.Annotations[lock] = ""
		c.CoreV1().ConfigMaps(namespace).Update(cfg)
		return true
	}
	return false
}
func (c *CassandraService) WriteReplaceIp(replaceIp string) error {
	// open file using WRITE & CREATE permission
	file, err := os.OpenFile(replaceFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(replaceIp)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}
	log.Infof("replace ip address updated to %s with %s", replaceFile, replaceIp)
	return nil
}
func (c *CassandraService) CreateFile() error {
	_, err := os.Stat(replaceFile)
	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(replaceFile)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}
func (c *CassandraService) WaitforReplacement(duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout while waiting for %s to be registered", ipAddress)
		case <-tick:
			if c.NewIpRegistered() {
				return nil
			}
		}
	}
}
func (c *CassandraService) NewIpRegistered() bool {
	nodetool := NewNodetool()
	status, err := nodetool.Status()
	if err != nil {
		log.Infof("\n\n nodetool error: %+v\n", err)
		return false
	}
	log.Infof("\n\n nodetool status: %+v\n", status)
	for _, dc := range status.Datacenters {
		for _, node := range dc.Nodes {
			if node.Address == ipAddress {
				if strings.Contains(node.State, "U") {
					return true
				}
				return false
			}
		}
	}
	return false
}
func (c *CassandraService) UpdateCM() (bool, error) {
	c.AquireLock()
	defer c.ReleaseLock()
	if !c.HasLock() {
		return false, fmt.Errorf("pod %s doesn't have the lock", pod)
	}
	cfg, err := c.CoreV1().ConfigMaps(namespace).Get(configmap, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warnf("configmap cassandra-topology configmap %s cannot be found...", configmap)
	}

	cfg.Data[pod] = ipAddress
	_, err = c.CoreV1().ConfigMaps(namespace).Update(cfg)
	if err != nil {
		return false, err
	}
	return true, err
}
func (c *CassandraService) Wait() {
	// monitor if the new node joined as UJ or UN with new ip address
	// and update the ip address in the configmap
	err := c.WaitforReplacement(4 * time.Minute)
	if err != nil {
		log.Errorf("error joining the cluster with replace ip: %v\n", err)
		os.Exit(1)
	} else {
		success, err := c.UpdateCM()
		if err != nil {
			log.Errorf("error updating the configmap with replace ip: %v\n", err)
			os.Exit(1)
		}
		if success {
			c.WriteReplaceIp("")
			log.Infoln("updated the configmap with replace ip")
			os.Exit(0)
		}
	}
}

func init() {
	//TODO don't use env variables ot fetch these vars, ideally a configmap mounted in pod with values
	namespace = os.Getenv("POD_NAMESPACE")
	pod = os.Getenv("POD_NAME")
	ipAddress = os.Getenv("POD_IP")
}
