#!/bin/bash

name=setu-checker
path=/go/src/github.com/sayuri567/setu-checker

build_go() {
    docker run --rm -v $GOPATH:/go golang:1.17.4 sh -c "cd $path; go mod tidy; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $name ./main.go"
    docker rmi $(docker images -f "dangling=true" -q)
    docker build -t setu-checker:1.0.0 --target setu-checker .
    docker-compose up -d
}

build_go
