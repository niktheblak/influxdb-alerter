#!/usr/bin/env bash

docker run \
  --rm \
  -v "$PWD/config.yaml:/etc/influxdb-alerter/config.yaml" \
  influxdb-alerter:latest
