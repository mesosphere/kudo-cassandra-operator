package suites

import "os"

func IsLocal() bool {
	return os.Getenv("BUILD_NUMBER") == ""
}

func SetLocalOnlyParameters(parameters map[string]string) {
	if IsLocal() {
		parameters["NODE_MEM_MIB"] = "768"
		parameters["NODE_MEM_LIMIT_MIB"] = "1024"
		parameters["NODE_CPU_MC"] = "1000"
		parameters["NODE_CPU_LIMIT_MC"] = "1500"
		parameters["PROMETHEUS_EXPORTER_ENABLED"] = "false"
	}
}
