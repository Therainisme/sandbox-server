#!/usr/bin/env bash

docker build --rm -t executor:v1 .
docker rmi $(docker images | grep "none" | awk '{print $3}')