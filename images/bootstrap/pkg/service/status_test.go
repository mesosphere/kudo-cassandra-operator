package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleDCStatus(t *testing.T) {

	statusContent := `
Datacenter: datacenter1
=======================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address     Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.244.2.6  232.33 KiB  256          66.7%             a444a8b8-4ffa-4148-9be9-b65ebde72ca5  rack1
UN  10.244.1.6  227.6 KiB  256          63.3%             08368dc2-a361-47f6-8c47-486e037037f6  rack1
UN  10.244.4.8  327.55 KiB  256          70.0%             7d256a00-3e00-4377-ae29-258b8aa5efd0  rack1
`
	status := ParseNodetoolStatus(statusContent)
	assert.NotEmpty(t, status.Datacenters)
	for _, dc := range status.Datacenters {
		assert.NotEmpty(t, dc.Nodes)
		assert.Equal(t, len(dc.Nodes), 3, "expected 3 nodes in nodetool status")
	}
}

func TestMultiDCStatus(t *testing.T) {

	statusContent := `
Datacenter: datacenter1
=======================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address     Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.244.2.6  232.33 KiB  256          66.7%             a444a8b8-4ffa-4148-9be9-b65ebde72ca5  rack1
UN  10.244.1.6  227.6 KiB  256          63.3%             08368dc2-a361-47f6-8c47-486e037037f6  rack1
UN  10.244.4.8  327.55 KiB  256          70.0%             7d256a00-3e00-4377-ae29-258b8aa5efd0  rack1
Datacenter: datacenter2
=======================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address     Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.244.2.6  232.33 KiB  256          66.7%             a444a8b8-4ffa-4148-9be9-b65ebde72ca5  rack1
UN  10.244.1.6  227.6 KiB  256          63.3%             08368dc2-a361-47f6-8c47-486e037037f6  rack1
UN  10.244.4.8  327.55 KiB  256          70.0%             7d256a00-3e00-4377-ae29-258b8aa5efd0  rack1
`
	status := ParseNodetoolStatus(statusContent)
	assert.NotEmpty(t, status.Datacenters)
	for _, dc := range status.Datacenters {
		assert.NotEmpty(t, dc.Nodes)
		assert.Equal(t, len(dc.Nodes), 3, "expected 3 nodes in nodetool status")
	}
}
