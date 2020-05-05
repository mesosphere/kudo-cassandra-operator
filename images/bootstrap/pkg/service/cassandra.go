package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

var (
	namespace      string
	pod            string
	configmap      string
	ipAddress      string
	bootstrapWait  string
	annotationLock = "cassandra.kudo.dev/annotationLock"
	replaceFile    = "/var/lib/cassandra/replace.ip"
)

type CassandraService struct {
	CMService *ConfigMapLock
}

func NewCassandraService(client *kubernetes.Clientset) *CassandraService {
	return &CassandraService{
		CMService: &ConfigMapLock{client},
	}
}

func (c *CassandraService) SetReplaceIP() (bool, error) {
	cfg, err := c.CMService.GetConfigMap(namespace, configmap)
	if errors.IsNotFound(err) {
		log.Errorf("bootstrap: configmap cassandra-topology configmap %s could not be found\n", configmap)
		return false, err
	}

	nodeIp := cfg.Data[pod]
	if nodeIp == "" {
		// first time bootstrap
		_, err = c.CMService.UpdateCM()
		return false, err
	} else if nodeIp != ipAddress {
		// new internal ip address
		if isBootstrapped() {
			// bootstrapped node needs no replace ip flag
			return true, nil
		} else {
			// node not bootstrapped and has an old ip address
			return true, c.WriteReplaceIp(nodeIp)
		}
	}
	return false, nil
}

func isBootstrapped() bool {
	// if cassandra is already bootstrapped the data/system dir is not empty
	_, err := os.Stat("/var/lib/cassandra/data/system")
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Errorf("bootstrap: error when checking for /var/lib/cassandra/data/system %v", err)
		return false
	}
	return true
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
	log.Infof("bootstrap: replace ip address updated to %s with %s", replaceFile, replaceIp)
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
	tick := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout while waiting for %s to be registered", ipAddress)
		case <-tick.C:
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
		log.Infof("bootstrap: nodetool error: %+v\n", err)
		return false
	}
	for _, dc := range status.Datacenters {
		for _, node := range dc.Nodes {
			if node.Address == ipAddress {
				return strings.Contains(node.State, "U")
			}
		}
	}
	return false
}

func (c *CassandraService) Wait() error {
	// monitor if the new node joined as UJ or UN with new ip address
	// and update the ip address in the configmap
	wait, err := time.ParseDuration(bootstrapWait)
	if err != nil {
		log.Errorf("bootstrap: missing BOOTSTRAP_TIMEOUT param: %v\n", err)
		return err
	}
	err = c.WaitforReplacement(wait * time.Minute)
	// re-joining can take really long time depending on the data
	if err != nil {
		log.Errorf("bootstrap: error joining the cluster with replace ip: %v\n", err)
		return err
	} else {
		success, err := c.CMService.UpdateCM()
		if err != nil {
			log.Errorf("bootstrap: error updating the configmap with replace ip: %v\n", err)
			return err
		}
		if success {
			log.Infoln("bootstrap: updating the configmap with replace ip")
			return c.WriteReplaceIp("")
		}
	}
	return nil
}

func init() {
	namespace = os.Getenv("POD_NAMESPACE")
	pod = os.Getenv("POD_NAME")
	ipAddress = os.Getenv("POD_IP")
	configmap = os.Getenv("CASSANDRA_IP_LOCK_CM")
	bootstrapWait = os.Getenv("BOOTSTRAP_TIMEOUT")
}
