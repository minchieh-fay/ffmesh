#!/bin/bash

name=ffmesh

DIST_DIR=dist

# linux-arm64
GOOS=linux GOARCH=arm64 go build -o ${DIST_DIR}/${name}-linux-arm64

# linux-amd64
GOOS=linux GOARCH=amd64 go build -o ${DIST_DIR}/${name}-linux-amd64

scp -P 2017 ${DIST_DIR}/${name}-linux-arm64 root@www.feyon.vip:/ff/tmp/${name}
scp -P 22 ${DIST_DIR}/${name}-linux-amd64 root@112.124.67.12:/root/ff/${name}
scp -P 22 ${DIST_DIR}/${name}-linux-amd64 root@8.219.218.174:/root/tmp/${name}

scp -P 2017 ${DIST_DIR}/ff.yaml root@www.feyon.vip:/ff/tmp/config.yaml
scp -P 22 ${DIST_DIR}/112.yaml root@112.124.67.12:/root/ff/config.yaml
scp -P 22 ${DIST_DIR}/8.yaml root@8.219.218.174:/root/tmp/config.yaml
