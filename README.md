# Sauron
This repo is for golang projects of straas

## Prerequisites
1. Golang 1.6+ - https://golang.org/dl/
2. Glide - https://github.com/Masterminds/glide
3. Add project root to GOPATH environment variable

## Installation
```
cd src/
glide install

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
