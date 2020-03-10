package cassandra

import (
	"gopkg.in/yaml.v2"
)

type NodeTopologyItem struct {
	Datacenter string `yaml:"datacenter"`
	Rack string `yaml:"rack"`
	Nodes int `yaml:"nodes"`
}

type NodeTopology []NodeTopologyItem

func (n NodeTopology) ToYAML() (string, error) {
	outputBytes, err := yaml.Marshal(n)

	return string(outputBytes), err
}
