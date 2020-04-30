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
	replaced, err := cassandraService.SetReplaceIP()
	if err != nil {
		log.Errorf("could not run the cassandra bootstrap: %v\n", err)
		os.Exit(1)
	}

	if replaced && cassandraService.Wait() != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
