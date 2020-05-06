package main

import (
	"os"

	"github.com/mesosphere/kudo-cassandra-operator/images/bootstrap/pkg/client"
	"github.com/mesosphere/kudo-cassandra-operator/images/bootstrap/pkg/service"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Infoln("Bootstrapping Cassandra...")
	client, err := client.GetKubernetesClient()
	if err != nil {
		log.Fatalf("Error initializing client: %+v", err)
	}
	cassandraService := service.NewCassandraService(client)

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "wait":
			if err := cassandraService.Wait(); err != nil {
				log.Errorf("timeout waiting for UN/UJ for Cassandra node: %v\n", err)
				os.Exit(1)
			}
		case "init":
			// bootstrap to fetch the replace IP
			if err := cassandraService.SetReplaceIPWithRetry(); err != nil {
				log.Errorf("could not run the cassandra bootstrap: %v\n", err)
				os.Exit(1)
			}
		default:
			log.Errorf("unrecognized command for cassandra bootstrap")
			os.Exit(1)
		}
	}
	os.Exit(0)
}
