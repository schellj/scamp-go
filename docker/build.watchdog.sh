#!/bin/sh
set -ex

docker build -t gcr.io/retailops-1/scamp-watchdog:dev -f docker/Dockerfile.watchdog .

if [[ $1 == "--push" ]]; then
  gcloud docker push gcr.io/retailops-1/scamp-watchdog:dev
fi