# Sauron
Sauron is a general purpose job engine

## Run

manually
```
PROJECT_ROOT=<Sauron project root>
cd ${PROJECT_ROOT}/src/straas.io/sauron/main
go run main.go \
  -envs=straas-staging \
  -dryRun=true \
  -esHosts=http://104.155.238.191:9200 \
  -configRoot=${PROJECT_ROOT}/config/sauron
```

dryrun script
```
PROJECT_ROOT=<Sauron project root>
cd ${PROJECT_ROOT}
bash script/sauron/run-dryrun.sh straas-staging
```

deploy docker script
```
PROJECT_ROOT=<Sauron project root>
cd ${PROJECT_ROOT}
bash script/sauron/deploy-k8s.sh <ENV:staging|production> <INIT:true|false> $TAG
```

## TODO:
1. Presistent storage
2. Handle query fail too many times(UNKNOWN)
