package main

import (
	"os"

	"github.com/mesosphere/kudo-cassandra-operator/images/bootstrap/pkg/client"
	"github.com/mesosphere/kudo-cassandra-operator/images/bootstrap/pkg/service"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Infoln("bootstrap: Bootstrapping Cassandra...")
	client, err := client.GetKubernetesClient()
	if err != nil {
		log.Fatalf("bootstrap: Error initializing client: %+v", err)
	}
	cassandraService := service.NewCassandraService(client)

	if len(os.Args) != 2 {
		log.Errorf("bootstrap: Wrong number of arguments: %d, must be 2: %v", len(os.Args), os.Args)
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "wait":
		log.Infof("Start waiting for Cassandra to be up")
		if err := cassandraService.Wait(); err != nil {
			log.Errorf("bootstrap: timeout waiting for UN/UJ for Cassandra node: %v\n", err)
			os.Exit(1)
		}
	case "init":
		// bootstrap to fetch the replace IP
		if err := cassandraService.SetReplaceIPWithRetry(); err != nil {
			log.Errorf("bootstrap: could not run the cassandra bootstrap: %v\n", err)
			os.Exit(1)
		}
		log.Infof("bootstrap: Finish Cassandra bootstrap init")
	default:
		log.Errorf("bootstrap: unrecognized command '%s' for cassandra bootstrap", command)
		os.Exit(1)
	}

	os.Exit(0)
}
