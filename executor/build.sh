#!/usr/bin/env bash

docker build -t therainisme/executor:2.0 .
docker rmi $(docker images | grep "none" | awk '{print $3}')