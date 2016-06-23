#!/bin/bash
# usage bash script/sauron/build-docker.sh <DOCKER_TAG>
# note that this script MUST run in project root

GIT_COMMIT=$(git rev-parse --short HEAD)
TAG=${1:-$GIT_COMMIT}

IMG_TAG=sauron:${TAG}
IMG_LATEST=sauron:latest
BUILD_IMG=build-${IMG_TAG}
CONTAINER_NAME=sauron-build-container
SOURCE_TAR=source.tar.gz
BUILD_FOLDER=build
SCRIPT_FOLDER=script/sauron

function clean_up {
  docker rm -f ${CONTAINER_NAME}
  # remove built docker
  docker rmi ${BUILD_IMG}
  # delete dangling dockers(<NONE>)
  docker rmi $(docker images -f "dangling=true" -q)
}

function safe_exec {
  echo CMD: $@
  $@
  if [ $? -ne 0 ]; then
    echo "$@ failed"
    # force cleanup if execute fail
    clean_up
    exit 1
  fi
}

function build_and_run_builder {  
  rm -f ${SOURCE_TAR}
  # archive source code only for clean build environment
  # once vendor is empty, tar not return 0
  # since we del tar file first, docker cp can test tar file existance
  tar zcvf ${SOURCE_TAR} config/sauron script/sauron src/straas.io/ src/glide.yaml

  safe_exec docker build -t ${BUILD_IMG} -f ${SCRIPT_FOLDER}/Dockerfile.build .
  safe_exec docker run -d --name ${CONTAINER_NAME} ${BUILD_IMG} tail -f /dev/null

  safe_exec docker cp ${SOURCE_TAR} ${CONTAINER_NAME}:/go/
  safe_exec docker exec -i ${CONTAINER_NAME} tar zxvf ${SOURCE_TAR}
  safe_exec docker exec -i ${CONTAINER_NAME} bash ${SCRIPT_FOLDER}/build-bin.sh
}

function run_test {
  safe_exec docker exec -it ${CONTAINER_NAME} go test $(go list straas.io/...)
}

# TODO: not used in current bash
function push_docker {
  PROJECT_ID=$1
  IMAGE_PATH=gcr.io/${PROJECT_ID}/${IMG_TAG}
  safe_exec docker tag ${IMG_TAG} ${IMAGE_PATH}
  safe_exec gcloud docker --project ${PROJECT_ID} push ${IMAGE_PATH}
  safe_exec docker rmi ${IMAGE_PATH}

  IMAGE_PATH=gcr.io/${PROJECT_ID}/${IMG_LATEST}
  safe_exec docker tag ${IMG_TAG} ${IMAGE_PATH}
  safe_exec gcloud docker --project ${PROJECT_ID} push ${IMAGE_PATH}
  safe_exec docker rmi ${IMAGE_PATH}  
}

# create build folder
rm -rf ${BUILD_FOLDER}
mkdir ${BUILD_FOLDER}

# build and run a docker for build code
build_and_run_builder

# run unit test
run_test

# copy necessary files to build folder
safe_exec docker cp ${CONTAINER_NAME}:/go/bin/sauron ${BUILD_FOLDER}/sauron

# build run docker
safe_exec docker build -t ${IMG_TAG} -f ${SCRIPT_FOLDER}/Dockerfile.run .

# push images to registry
push_docker ikala-infra

clean_up
