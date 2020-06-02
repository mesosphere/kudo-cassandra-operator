
1. assert kudo 
    k kudo init --wait
    k kudo init --wait --unsafe-self-signed-webhook-ca

2. create ns
2. install CR
    assert
3. install with params
    with instance name
    assert

## Prerequisites

* Kubernetes cluster up
* KUDO CLI installed on node kuttl is running from
* KUDO manager installed in the cluster


# TestSuites

* **Setup:** used to setup a MWT test use the following command 

`kubectl kuttl test setup/  --parallel 1 --skip-delete`

The `parallel 1` is important for setup, as the order of the folders and the order of there execution matters.  The tests are designed to allow for a restart.  If kudo confirmation is guaranteed, it is possible to run: `kubectl kuttl test setup/ --parallel 1 --skip-delete --test 01-cassandra-install`.  The tests inside "cassandra-install" are also allowed to run multiple tests with good end results.  This is the reason command failures are ignored.  If a command fails we don't care unless the desired asserted state is not reached (which is built into the test).

* **Teardown:** used to remove cassandra (and verify it has been removed)

`kuttl test teardown/ `

## Tests under **Setup**

* 00-kudo-check: confirms
    1. commands locally are present
    2. kudo manager is ready at the server

* 01-cassandra-install
    1. first creates and asserts "cassandra" namespace exists
    2. installs the cassandra CR and asserts they exist
    3. installs cassandra with parameters and asserts that the deployment plan is complete
    4. runs the nodetool on node 0 for output to stdout


**NOTES:** 
1. The MWT parameter values need to be changed. This was built and tested on konvoy but not on a MWT env
2. The timeouts may need to change.  In particular the wait for namespace, the wait for deploy to finish and the wait for deletes.
3. This was tested with konvoy with config setting established with `./konvoy apply kubeconfig --force-overwrite`