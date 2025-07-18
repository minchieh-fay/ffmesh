#!/bin/bash

name=ffmesh

DIST_DIR=dist

# linux-arm64
GOOS=linux GOARCH=arm64 go build -o ${DIST_DIR}/${name}-linux-arm64

# linux-amd64
GOOS=linux GOARCH=amd64 go build -o ${DIST_DIR}/${name}-linux-amd64
