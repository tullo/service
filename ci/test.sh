set -e

go test -mod=vendor -failfast -short -test.timeout=90s -run="^Test"  \
    ./app/... ./business/... ./foundation/...
