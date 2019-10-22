# Monitoring (Using Prometheus Service Monitor)

KUDO Cassandra operator comes with criteo's [cassandra exporter](https://github.com/criteo/cassandra_exporter), which exports Cassandra metrics throught a prometheus friendly endpoint.

To use the prometheus service monitor, its necessary to have installed the prometheus operator previously in the cluster.

When Cassandra operator deployed with parameter `PROMETHEUS_EXPORTER_ENABLED=true` (which defaults to `true`) then:

- Each Pod will be added with `prometheus-exporter` container which will export metrics at parameter `PROMETHEUS_EXPORTER_PORT`, by default its set to `7200`
- Adds a port named `prometheus-exporter-port` to the Cassandra Service
- Adds a label `kudo.dev/servicemonitor: "true"` for the service monitor discovery.

```
kubectl describe svc cassandra-svc
...
Port:              prometheus-exporter-port  7200/TCP
TargetPort:        7200/TCP
...
```

If the Cassandra Service in the default namespace, we will need to use the next example of the service-monitor to see metrics in Prometheus.

```
kubectl create -f resources/service-monitor.yaml
```
