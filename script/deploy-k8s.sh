#!/bin/bash
# use temporarly, script will be replaced by kalajan
# following script is copied from revealer deploy script

RUN_ENV=$1
TAG=$2
INIT=$3

CPU=200m
MEM=200Mi

REGION="asia-east1"
ZONE="asia-east1-b"
RC_NAME="sauron"

# make a temp file for deploy file
DEPLOY_FILE=$(mktemp $TMPDIR/$(uuidgen).yaml)
TEMPLATE_FILE="k8s/rc-template.yaml"

function safeExec {
  echo CMD: $@
  $@
  if [ $? -ne 0 ]; then
    echo "$@ failed"
    cleanup
    exit 1
  fi
}

function cleanup {
  echo "perform cleanup"
  rm ${DEPLOY_FILE}
}

function escape {
  echo $(echo $1 | sed 's/\//\\\//g')
}

case "${RUN_ENV}" in
"straas-staging")
  PROJECT_ID="straasio-staging"
  CLUSTER_NAME=$PROJECT_ID
  ES_HOSTS=http://infra-elasticsearch-straas-staging-1:9200
  ES_HOSTS=${ES_HOSTS},http://infra-elasticsearch-straas-staging-2:9200
  ;;
"straas-production")
  PROJECT_ID="straasio-production"
  CLUSTER_NAME=$PROJECT_ID
  ES_HOSTS=http://infra-elasticsearch-straas-production-1:9200
  ES_HOSTS=${ES_HOSTS},http://infra-elasticsearch-straas-production-2:9200
  ES_HOSTS=${ES_HOSTS},http://infra-elasticsearch-straas-production-3:9200
  ;;
*)
  echo "Unsupported environment. Terminating..."
  exit 1
  ;;
esac

IMAGE=gcr.io/${PROJECT_ID}/sauron:${TAG}

echo "Start to deploy ${IMAGE} to ${RUN_ENV} ..."

# escape eshosts forward slash for sed replacement
IMAGE=$(escape "${IMAGE}")
ES_HOSTS=$(escape "${ES_HOSTS}")

cp ${TEMPLATE_FILE} ${DEPLOY_FILE}
sed -i '' 's/${RC_NAME}/'${RC_NAME}'/g' ${DEPLOY_FILE}
sed -i '' 's/${IMAGE}/'${IMAGE}'/g' ${DEPLOY_FILE}
sed -i '' 's/${RUN_ENV}/'${RUN_ENV}'/g' ${DEPLOY_FILE}
sed -i '' 's/${CPU}/'${CPU}'/g' ${DEPLOY_FILE}
sed -i '' 's/${MEM}/'${MEM}'/g' ${DEPLOY_FILE}
sed -i '' 's/${ES_HOSTS}/'${ES_HOSTS}'/g' ${DEPLOY_FILE}

# Set KUBECTL to command full path if such commmand exists; otherwise,
# set KUBECTL as absolute path so that it can run on Jenkins build node.
if hash kubectl 2> /dev/null; then
  KUBECTL=$(which kubectl)
else
  KUBECTL="/home/jenkins/google-cloud-sdk/bin/kubectl"
fi

safeExec gcloud config set compute/region $REGION
safeExec gcloud config set compute/zone $ZONE
safeExec gcloud config set project $PROJECT_ID

# only set service account when the user is not in GCE
#if [[ $(curl -s "http://metadata/computeMetadata/v1/project/project-id" -H "Metadata-Flavor: Google") != $PROJECT_ID ]]; then
#  safeExec gcloud config set account gcpservieaccount@$PROJECT_ID.iam.gserviceaccount.com
#fi
if [[ $INIT != "init" ]]; then
  safeExec gcloud config set container/cluster $CLUSTER_NAME
  safeExec gcloud container clusters get-credentials $CLUSTER_NAME
fi

$KUBECTL get rc ${RC_NAME} && safeExec $KUBECTL delete rc ${RC_NAME}
safeExec $KUBECTL create -f $DEPLOY_FILE

# cleanup files and resources
cleanup
