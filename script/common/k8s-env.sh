#!/bin/bash

PROJECT_ID="ikala-infra"
CLUSTER_NAME="${PROJECT_ID}-k8s"
REGION="asia-east1"
ZONE="asia-east1-b"

function safeExec {
  echo "$@"
  "$@"
  if [ $? -ne 0 ]; then
    echo "$@ failed"
    exit 1
  fi
}

function k8s_create {
  $KUBECTL create -f -
}

function k8s_delete_rc {
  RC_NAME=$1
  $KUBECTL get rc ${RC_NAME} && $KUBECTL delete rc ${RC_NAME}
}

# Set KUBECTL to command full path if such commmand exists; otherwise,
# set KUBECTL as absolute path so that it can run on Jenkins build node.
if hash kubectl 2> /dev/null; then
  KUBECTL=$(which kubectl)
else
  KUBECTL="/home/jenkins/google-cloud-sdk/bin/kubectl"
fi

# always init for simplify
safeExec gcloud config set compute/region $REGION
safeExec gcloud config set compute/zone $ZONE
safeExec gcloud config set project $PROJECT_ID
safeExec gcloud config set container/cluster $CLUSTER_NAME
safeExec gcloud container clusters get-credentials $CLUSTER_NAME
