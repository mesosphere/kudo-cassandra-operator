FROM golang:1.14.2-stretch AS bootstrap-bin
ADD ./bootstrap /bootstrap
WORKDIR /bootstrap
RUN  apt-get update && apt-get install ca-certificates
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap

FROM cassandra:3.11.6

RUN set -eux; \
  apt-get update; \
  apt-get install -y --no-install-recommends libcap2-bin; \
    echo "/opt/java/openjdk/lib/amd64/" >> /etc/ld.so.conf.d/java.conf; \
    echo "/opt/java/openjdk/lib/amd64/jli" >> /etc/ld.so.conf.d/java.conf; \
    ldconfig; \
  # Setting this capability will cause running the Java binary to not be
  # "recognized" when the Cassandra binary is run in a container without
  # IPC_LOCK capabilities, showing a message like "Cassandra 3.0 and later
  # require Java 8u40 or later". This is because even though the binary in the
  # Docker image will have the capabilities, the container execution environment
  # might not have them.
  #
  # Also, capabilities need to be added to the actual Java binary, not the
  # Cassandra script.
  setcap cap_sys_resource,cap_ipc_lock+eip $(readlink -f $(which java))

USER cassandra
COPY --from=bootstrap-bin /bootstrap/bootstrap /etc/cassandra-bootstrap/
