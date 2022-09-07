#!/bin/bash
rm -rf  staging
go build -o  staging/helloworld -ldflags="-X 'main.Version=$1'"