# Sauron
This repo is for golang projects of straas

## Prerequisites
1. Golang 1.6+ - https://golang.org/dl/
2. Glide - https://github.com/Masterminds/glide
3. Add project root to GOPATH environment variable
4. make sure $GOPATH/bin in $PATH
5. install https://github.com/golang/lint

# Dev docker (Recommended)
```
// in project root
bash script/common/docker-dev.sh build
bash script/common/docker-dev.sh run
```

## Installation
```
cd src/
glide install

```

## Git hook
```
// in project root
ln -s ../../.git-hooks/pre-push .git/hooks/pre-push
```

## linter
```
go vet $(go list straas.io/...)
go list straas.io/... | grep -v "mocks" | xargs -n 1 golint
```

## testing
```
go test -cover $(go list straas.io/...)
```
