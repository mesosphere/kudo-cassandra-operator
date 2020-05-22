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

func TestGossipActive(t *testing.T) {

	infoContent := `
ID                     : 6ca2e4cf-f289-4447-8c93-773a246abfcd
Gossip active          : true
Thrift active          : false
Native Transport active: true
Load                   : 108.64 KiB
Generation No          : 1589987878
Uptime (seconds)       : 351
Heap Memory (MB)       : 196.84 / 460.81
Off Heap Memory (MB)   : 0.00
Data Center            : datacenter1
Rack                   : rack1
Exceptions             : 0
Key Cache              : entries 11, size 896 bytes, capacity 23 MiB, 90 hits, 110 requests, 0.818 recent hit rate, 14400 save period in seconds
Row Cache              : entries 0, size 0 bytes, capacity 0 bytes, 0 hits, 0 requests, NaN recent hit rate, 0 save period in seconds
Counter Cache          : entries 0, size 0 bytes, capacity 11 MiB, 0 hits, 0 requests, NaN recent hit rate, 7200 save period in seconds
Chunk Cache            : entries 18, size 1.12 MiB, capacity 83 MiB, 29 misses, 189 requests, 0.847 recent hit rate, NaN microseconds miss latency
Percent Repaired       : 100.0%
Token                  : (invoke with -T/--tokens to see all 256 tokens)
`

	gossipActive, err := parseInfoGossipStatus(infoContent)
	assert.Nil(t, err)
	assert.True(t, gossipActive)
}

func TestGossipInactive(t *testing.T) {

	infoContent := `
ID                     : 6ca2e4cf-f289-4447-8c93-773a246abfcd
Gossip active          : false
Thrift active          : false
Native Transport active: true
Load                   : 108.64 KiB
Generation No          : 1589987878
Uptime (seconds)       : 351
Heap Memory (MB)       : 196.84 / 460.81
Off Heap Memory (MB)   : 0.00
Data Center            : datacenter1
Rack                   : rack1
Exceptions             : 0
Key Cache              : entries 11, size 896 bytes, capacity 23 MiB, 90 hits, 110 requests, 0.818 recent hit rate, 14400 save period in seconds
Row Cache              : entries 0, size 0 bytes, capacity 0 bytes, 0 hits, 0 requests, NaN recent hit rate, 0 save period in seconds
Counter Cache          : entries 0, size 0 bytes, capacity 11 MiB, 0 hits, 0 requests, NaN recent hit rate, 7200 save period in seconds
Chunk Cache            : entries 18, size 1.12 MiB, capacity 83 MiB, 29 misses, 189 requests, 0.847 recent hit rate, NaN microseconds miss latency
Percent Repaired       : 100.0%
Token                  : (invoke with -T/--tokens to see all 256 tokens)
`

	gossipActive, err := parseInfoGossipStatus(infoContent)
	assert.Nil(t, err)
	assert.False(t, gossipActive)
}
