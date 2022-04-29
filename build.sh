#!/usr/bin/env bash

docker build --rm -t therainisme/sandbox-server:2.1 .
docker rmi $(docker images | grep "none" | awk '{print $3}')