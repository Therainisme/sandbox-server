#!/usr/bin/env bash

docker build --rm -t therainisme/executor:1.0 .
docker rmi $(docker images | grep "none" | awk '{print $3}')