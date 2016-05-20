#!/bin/bash

# NOTE: this script must run in project ROOT
# install pkgs
cd src/
glide install
cd -

# build sauron config builder
go install straas.io/sauron/main
