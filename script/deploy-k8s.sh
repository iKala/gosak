#!/bin/bash
# use temporarly, script will be replaced by kalajan
# following script is copied from revealer deploy script

source script/k8s-env.sh

GIT_COMMIT=$(git rev-parse --short HEAD)

RUN_ENV_GROUP=$1
INIT=$2
TAG=${3:-$GIT_COMMIT}

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
          env_group: ${RUN_ENV_GROUP}
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

# external ip is for stackdriver uptime check
# source ip of firewall rule is copied from uptime check options button
function service_cfg {
  cat <<-EOF
  apiVersion: v1
  kind: Service
  metadata:
    name: sauron-production
    labels:
      app: sauron
      env_group: production
  spec:
    ports:
    - port: 8000
      targetPort: 8000
    selector:
      app: sauron
      env_group: production
    type: LoadBalancer
    loadBalancerIP: 104.199.151.71
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

if [[ "$INIT" == "true" ]]; then
  service_cfg | k8s_create || exit 1
fi

IMAGE=gcr.io/${PROJECT_ID}/sauron:${TAG}
CPU=200m
MEM=200Mi

echo "Start to deploy ${IMAGE} to ${RUN_ENV} ..."
k8s_delete_rc ${RC_NAME}
rc_cfg | k8s_create || exit 1
