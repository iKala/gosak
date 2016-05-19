#!/bin/bash
# usage: bash script/run-dryrun.sh straas-production <IMG>
# note that this script MUST run in project root

RUN_ENV=$1
DEFAULTIMG=gcr.io/straasio-staging/sauron:latest
IMAGE=${2:-$DEFAULTIMG}

# prepare variables for different env
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
  -v $(pwd)/config/sauron:/configForDryrun \
  ${IMAGE} \
  -dryRun=true \
  -configRoot=/configForDryrun \
  -logLevel=error \
  -esHosts=${ES_HOSTS} \
  -envs=${RUN_ENV} $@
