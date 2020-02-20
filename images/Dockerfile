FROM cassandra:3.11.5

RUN set -eux; \
  rm /etc/apt/sources.list.d/cassandra.list; \
  apt-get update; \
  apt-get install -y --no-install-recommends libcap2-bin; \
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
