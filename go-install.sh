#!/bin/bash

echo "Installing Go release '$1'"
wget --no-verbose "https://go.dev/dl/go$1.linux-amd64.tar.gz"

# Remove any previous Go installation and extract tar file
rm -rf /usr/local/go && tar -C /usr/local -xzf go$1.linux-amd64.tar.gz
rm go$1.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin
go version
