#!/bin/bash
# usage: bash script/run-docker.sh straas-production <IMG>
# note that this script MUST run in project root

RUN_ENV=$1
IMAGE=$2

# prepare variables for different env
case "$RUN_ENV" in
  "straas-production")
    ES_HOSTS=http://infra-elasticsearch-straas-production-1:9200
    ES_HOSTS=${ES_HOSTS},http://infra-elasticsearch-straas-production-2:9200
    ES_HOSTS=${ES_HOSTS},http://infra-elasticsearch-straas-production-3:9200
  ;;
  "straas-production")
    ES_HOSTS=http://infra-elasticsearch-straas-staging-1:9200
    ES_HOSTS=${ES_HOSTS},http://infra-elasticsearch-straas-staging-2:9200
  ;;
  "lh-production,lh-alpha" | "lh-production" | "lh-alpha")
    ES_HOSTS=http://zebra-elastic-a:9200
    ES_HOSTS=${ES_HOSTS},http://zebra-elastic-b:9200
    ES_HOSTS=${ES_HOSTS},http://zebra-elastic-c:9200
  ;;
  *)
    echo "no such project ${PROJECT}"
    exit 1
  ;;
esac

docker run -d \
  ${IMAGE} \
  -configRoot=/config \
  -esHosts=${ES_HOSTS} \
  -envs=${RUN_ENV} \
  -dryRun=false
