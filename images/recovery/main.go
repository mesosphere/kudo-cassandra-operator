package main

import (
	"context"

	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/client"
	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/controller"

	"log"
)

func main() {

	log.Printf("bootstrapping cassandra recovery controller...")

	clientSet, err := client.GetKubeClient()
	if err != nil {
		log.Fatalf("failed to get kube client: %v", err)
		return
	}

	cont := controller.NewController(clientSet, controller.NewOptions())
	cont.Run(context.Background())
}
