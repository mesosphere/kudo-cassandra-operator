package cassandra

import (
	"gopkg.in/yaml.v2"
)

type TopologyDatacenterItem struct {
	// The name of the datacenter as known to cassandra
	Datacenter string `yaml:"datacenter"`

	// Node selector labels
	DatacenterLabels map[string]string `yaml:"datacenterLabels"`

	// They label key on nodes that distinguishes different racks
	RackLabelKey string `yaml:"rackLabelKey"`

	// The different racks in this datacenter
	Racks []TopologyRackItem `yaml:"racks"`

	// The total number of nodes in the datacenter, will get distributed over all racks
	Nodes int `yaml:"nodes"`
}

type TopologyRackItem struct {
	// The name of the rack, as known to cassandra
	Rack string `yaml:"rack"`

	// The values of `rackLabelKey` that make up this rack
	RackLabelValue string `yaml:"rackLabelValue"`
}

type NodeTopology []TopologyDatacenterItem

func (n NodeTopology) ToYAML() (string, error) {
	outputBytes, err := yaml.Marshal(n)

	return string(outputBytes), err
}

func TopologyFromYaml(str string) (NodeTopology, error) {
	topo := NodeTopology{}
	err := yaml.Unmarshal([]byte(str), &topo)
	return topo, err
}
