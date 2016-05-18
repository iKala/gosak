#!/bin/bash
# usage: bash run-dryrun.sh straas-production <IMG>

RUN_ENV=$1
DEFAULTIMG=gcr.io/straasio-staging/sauron:latest
IMAGE=${2:-$DEFAULTIMG}

# run bosun
case "$RUN_ENV" in
  "straas-production")
    ES_HOSTS=http://104.155.232.6:9200
  ;;
  "straas-staging")
    ES_HOSTS=http://104.155.238.191:9200
  ;;
  "lh-production")
    ES_HOSTS=http://104.155.229.57:9200
  ;;
  "lh-alpha")
    ES_HOSTS=http://104.155.229.57:9200
  ;;
  *)
    echo "no such project ${PROJECT}"
    exit 1
  ;;
esac

gcloud docker --project straasio-staging pull $IMAGE
docker run -it --rm \
  -v $(pwd)/config:/configForDryrun \
  ${IMAGE} \
  -dryRun \
  -configRoot=/configForDryrun \
  -logLevel=error \
  -esHosts=${ES_HOSTS} \
  -envs=${RUN_ENV} $@
