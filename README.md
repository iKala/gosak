# Sauron
Sauron is a general purpose job engine

## Prerequisites
1. Golang 1.6+ - https://golang.org/dl/
2. Glide - https://github.com/Masterminds/glide
3. Add project root to GOPATH environment variable

## Installation
```
cd src/
glide install

```

## File layout
1. src/straas.io/sauron  sauron job engne
2. src/straas.io/base common library

## Run
```
PROJECT_ROOT=<Sauron project root>
cd ${PROJECT_ROOT}/src/straas.io/sauron/main
go run main.go \
  -envs=straas-staging \
  -dryRun \
  -esHosts=http://104.155.238.191:9200 \
  -configRoot=${PROJECT_ROOT}/config \
```

## TODO:
1. Dryrun mode (a.k.a fake notification, but queriable)
2. Presistent storage
3. Push to private registry
4. Handle query fail too many times(UNKNOWN)
5. Better query func (e.g. lastfor)
