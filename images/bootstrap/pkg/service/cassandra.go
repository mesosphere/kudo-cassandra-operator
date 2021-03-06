package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	ANNOTATION_LOCK = "cassandra.kudo.dev/annotationLock"
	REPLACE_FILE    = "/var/lib/cassandra/replace.ip"
	RETRY_DELAY     = 3 * time.Second
	RETRY_ATTEMPTS  = 10
)

var (
	namespace                string
	podName                  string
	configmapName            string
	podIpAddress             string
	bootstrapWait            string
	jmxPort                  string
	useSSL                   bool
	shutdownOldReachableNode bool
)

type CassandraService struct {
	CMService *ConfigMapLock
}

func init() {
	namespace = os.Getenv("POD_NAMESPACE")
	podName = os.Getenv("POD_NAME")
	podIpAddress = os.Getenv("POD_IP")
	configmapName = os.Getenv("CASSANDRA_IP_LOCK_CM")
	bootstrapWait = os.Getenv("BOOTSTRAP_TIMEOUT")
	jmxPort = os.Getenv("JMX_PORT")
	useSSL = os.Getenv("USE_SSL") == "true"
	shutdownOldReachableNode = os.Getenv("SHUTDOWN_OLD_REACHABLE_NODE") == "true"
}

func NewCassandraService(client *kubernetes.Clientset) *CassandraService {
	return &CassandraService{
		CMService: &ConfigMapLock{client},
	}
}

func (c *CassandraService) SetReplaceIPWithRetry() error {
	return retry.Do(c.SetReplaceIP, retry.Delay(RETRY_DELAY), retry.Attempts(RETRY_ATTEMPTS))
}

func (c *CassandraService) SetReplaceIP() error {
	cfg, err := c.CMService.GetConfigMap(namespace, configmapName)
	if errors.IsNotFound(err) {
		log.Errorf("bootstrap: cassandra-topology configmap %s could not be found\n", configmapName)
		return err
	}
	oldIp := cfg.Data[podName]
	log.Infof("bootstrap: Got old IP %s for pod %s, current IP is %s", oldIp, podName, podIpAddress)
	if oldIp == podIpAddress || oldIp == "" {
		return nil
	}

	if shutdownOldReachableNode {
		// This is guarded by a feature flag, as this call can have quite a timeout and delay node startup
		if isOldNodeReachableAndUp(oldIp) {
			log.Infof("old node %s is still reachable and marked as UP. Try to shutdown old node now", oldIp)
			tryOldNodeShutdown(oldIp)
			return fmt.Errorf("tried to shutdown old node %s, wait for retry", oldIp)
		}
	}

	// new internal ip address
	if isBootstrapped() {
		log.Infof("bootstrap: Node is already bootstrapped, no need for replace IP")
		// bootstrapped node needs no replace ip flag
		return nil
	}

	log.Infof("bootstrap: Node is not bootstrapped, add replace ip to startup")
	// node not bootstrapped and has an old ip address
	return c.WriteReplaceIp(oldIp)
}

func isOldNodeReachableAndUp(oldIP string) bool {
	nt := NewRemoteNodetool(oldIP, jmxPort, useSSL)
	status, err := nt.Status()
	log.Infof("Old Node Status: %v (%v)", status, err)
	if err != nil {
		log.Info("Old node seems to be not reachable anymore")
		return false
	}
	if !status.HasUpNode(oldIP) {
		log.Info("Could connect to old node, but the node is not marked as UP anymore.")
		return false
	}
	gossipActive, err := nt.HasActiveGossip()
	if err != nil {
		log.Infof("Failed to get Gossip State: %v", err)
	}
	return gossipActive
}

// tryOldNodeShutdown tries to connect to the old node and shut it down. Returns true if it was possible to
// connect and false if the old node was not reachable
func tryOldNodeShutdown(oldIp string) {
	nt := NewRemoteNodetool(oldIp, jmxPort, useSSL)

	log.Infof("Try to drain old node...")
	if _, err := nt.RunCommand("drain"); err != nil {
		log.Errorf("Nodetool drain on remote host failed:%v", err)
	}

	// We actually would like to do "stopdaemon", which currently throws an exception. "disablegossip" works as well to remove the node from the cluster
	log.Infof("Try to stop old node...")
	if _, err := nt.RunCommand("disablegossip"); err != nil {
		log.Errorf("Nodetool disablegossip on remote host failed:%v", err)
	}
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
	file, err := os.OpenFile(REPLACE_FILE, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
	log.Infof("bootstrap: replace ip address updated to %s with %s", REPLACE_FILE, replaceIp)
	return nil
}

func (c *CassandraService) WaitforReplacement(duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout while waiting for %s to be registered", podIpAddress)
		case <-tick.C:
			if c.NewIpRegistered() {
				return nil
			}
		}
	}
}

func (c *CassandraService) NewIpRegistered() bool {
	nodetool := NewNodetool(useSSL)
	status, err := nodetool.Status()
	if err != nil {
		log.Infof("bootstrap: nodetool error: %+v\n", err)
		return false
	}
	for _, dc := range status.Datacenters {
		for _, node := range dc.Nodes {
			if node.Address == podIpAddress {
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
	}
	log.Infoln("bootstrap: updating the configmap with new node ip")
	if err := retry.Do(c.CMService.UpdateCM, retry.Delay(RETRY_DELAY), retry.Attempts(RETRY_ATTEMPTS)); err != nil {
		log.Errorf("bootstrap: error updating the configmap with replace ip: %v\n", err)
		return err
	}
	log.Infoln("bootstrap: reset replace ip")
	return c.WriteReplaceIp("")
}
