#!/bin/bash
# use temporarly, script will be replaced by kalajan
# following script is copied from revealer deploy script

source script/k8s-env.sh

GIT_COMMIT=$(git rev-parse --short HEAD)

RUN_ENV_GROUP=$1
TAG=${2:-$GIT_COMMIT}

function rc_cfg {
  cat <<-EOF
  apiVersion: v1
  kind: ReplicationController
  metadata:
    name: ${RC_NAME}
  spec:
    replicas: 1
    template:
      metadata:
        labels:
          app: sauron
      spec:
        containers:
        - name: sauron
          image: ${IMAGE}
          imagePullPolicy: Always
          args:          
            - -configRoot=/config
            - -envs=${RUN_ENV}
            - -esHosts=${ES_HOSTS}
            - -dryRun=false
          resources:
            limits:
              cpu: ${CPU}
              memory: ${MEM}
          livenessProbe:
            httpGet:
              port: 8000
            initialDelaySeconds: 10
EOF
}

case "${RUN_ENV_GROUP}" in
  "staging")
    RUN_ENV=straas-staging
    RC_NAME="sauron-staging"
    ES_HOSTS=http://elasticsearch-main-ikalainfra-staging-1:9200
    ES_HOSTS=${ES_HOSTS},http://elasticsearch-main-ikalainfra-staging-2:9200
  ;;
  "production")
    RUN_ENV=straas-production
    RC_NAME="sauron-production"
    ES_HOSTS=http://elasticsearch-main-ikalainfra-production-1:9200
    ES_HOSTS=${ES_HOSTS},http://elasticsearch-main-ikalainfra-production-2:9200
    ES_HOSTS=${ES_HOSTS},http://elasticsearch-main-ikalainfra-production-3:9200
  ;;
  *)
    echo "Unsupported environment. Terminating..."
    exit 1
  ;;
esac

echo "Start to deploy ${IMAGE} to ${RUN_ENV} ..."

IMAGE=gcr.io/${PROJECT_ID}/sauron:${TAG}
CPU=200m
MEM=200Mi

#k8s_delete_rc ${RC_NAME}
rc_cfg | k8s_create || exit 1
