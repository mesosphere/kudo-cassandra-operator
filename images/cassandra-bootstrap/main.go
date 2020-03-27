package main

import (
	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-bootstrap/pkg/client"
	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-bootstrap/pkg/service"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Infoln("Running cassandra bootstrap...")
	k8sClient, err := client.GetKubernetesClient()
	if err != nil {
		log.Fatalf("Error initializing client: %+v", err)
	}
	cassandraService := service.CassandraService{
		k8sClient,
	}

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "wait" {
			cassandraService.Wait()
			os.Exit(0)
		}
	}


	log.Infoln("Running cassandra ip detection...")
	replaced, err := cassandraService.SetReplaceIP()
	if err != nil {
		log.Errorf("could not run the cassandra bootstrap: %v\n", err)
	}

	if replaced {
		cassandraService.Wait()
	}
	os.Exit(0)

}
