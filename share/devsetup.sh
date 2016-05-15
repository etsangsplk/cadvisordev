#!/bin/bash

echo Installing build dependencies
apt-get update -y
apt-get install git -y
apt-get install build-essential -y
apt-get install curl -y
apt-get install docker.io -y

echo Installing Go 1.5.4
curl -O "https://storage.googleapis.com/golang/go1.5.4.linux-amd64.tar.gz"
tar -C /usr/local -xzf go1.5.4.linux-amd64.tar.gz

export WF_ADD_TAGS="az=\"us-west-2\" app=\"cadvisortesting\""
export WF_INTERVAL=10
export GOPATH=/opt/share && export PATH=$PATH:/usr/local/go/bin && export PATH=$PATH:$GOPATH/bin

echo Downloading cAdvisor codebase and dependencies
go get -d github.com/google/cadvisor
go get github.com/tools/godep
