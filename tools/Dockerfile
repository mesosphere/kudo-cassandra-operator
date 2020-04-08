FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y gettext-base git python3 python3-pip
RUN pip3 install pyyaml pytablewriter

WORKDIR /opt/kudo-cassandra-operator
