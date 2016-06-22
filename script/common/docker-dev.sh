#!/bin/bash

CMD=$1
DOCKERMACHINE=$2

IMAGE="straas/wego"
WORKDIR=/go
CONTAINER_NAME=wego-dev
DIR=script/common

case $CMD in
  "help")
    echo "usage:"
    echo "${DIR}/docker-dev.sh [CMD] [ARG...]"
    echo ""
    echo "available CMD"
    echo "help             show help messages"
    echo "build            build dev docker"
    echo "   virtualbox    [optional] build with virtualbox"
    echo "run              run dev docker"
    echo "start_daemon     run dev docker as daemon"
    echo "stop_daemon      stop the dev docker daemon"
    echo "exec             exec command in daemon. e.g. docker/docker-dev.sh exec gulp help"
    ;;
  "build")
    MY_UID=$(id -u)
    if [[ "$DOCKERMACHINE" == "virtualbox" ]]; then MY_UID=1000; fi
    docker build -t "$IMAGE" --build-arg MY_UID=$MY_UID -f ${DIR}/Dockerfile.dev ${DIR}
    ;;
  "run")
    docker run --rm -it -v $(pwd):${WORKDIR} "$IMAGE" /bin/bash
    ;;
  "start_daemon")
    docker run -d --name ${CONTAINER_NAME} -v `pwd`:${WORKDIR} "$IMAGE" tail -f /dev/null
    ;;
  "stop_daemon")
    docker rm -f ${CONTAINER_NAME}
    ;;
  "exec")
    shift
    docker exec -it ${CONTAINER_NAME} env TERM=xterm $@
    ;;
  *)
    echo "no such command $CMD, please run '${DIR}/docker-dev.sh help' for detail usage"
    exit 1
    ;;
esac
