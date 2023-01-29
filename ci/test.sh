#!/usr/bin/env bash
set -e
echo "" > coverprofile.cov

go test -short -mod=vendor -failfast -test.timeout=90s \
    -covermode=count -coverprofile=coverprofile.cov -run="^Test" \
    -coverpkg=$(go list -mod=vendor ./... | grep -v "/test" | tr '\n' ',') \
    ./...
# go tool cover -func=coverprofile.cov
# go tool cover -html=coverprofile.cov
