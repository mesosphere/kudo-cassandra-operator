package main

import (
	"context"

	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/client"
	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/controller"


	"log"
)

func main() {

	log.Printf("bootstrapping cassandra recovery POC...")

	client, _ := client.GetKubernetesClient()
	cont := controller.NewController(client)
	cont.Run(context.Background())
}
