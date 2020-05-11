package suites

import "os"

func IsLocalCluster() bool {
	return os.Getenv("LOCAL_CLUSTER") == "true"
}

// SetSuitesParameters adds a set of common parameters
// - parameters that should be set across all testing suites
// - parameters for local testing in a minikube or other restricted environments
// This includes limited CPU and memory settings as well as disabling the Prometheus exporter when running in local cluster
func SetSuitesParameters(parameters map[string]string) {
	parameters["NODE_DOCKER_IMAGE"] = os.Getenv("CASSANDRA_DOCKER_IMAGE")
	if IsLocalCluster() {
		parameters["NODE_MEM_MIB"] = "768"
		parameters["NODE_MEM_LIMIT_MIB"] = "1024"
		parameters["NODE_CPU_MC"] = "1000"
		parameters["NODE_CPU_LIMIT_MC"] = "1500"
		parameters["PROMETHEUS_EXPORTER_ENABLED"] = "false"
	}
}
