#!/bin/bash

export GOOS=linux
export GOARCH=amd64
go build -o cryp-kaspad-server ../cmd/cryp-kaspad/main.go
