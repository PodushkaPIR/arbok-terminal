#!/bin/bash
set -e

usage() {
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  build    Build the binary"
    echo "  test     Run tests"
    echo "  vet      Run go vet"
    echo "  fmt      Check formatting"
    echo "  lint     Run vet + fmt check"
    echo "  all      Run all checks (vet + fmt + test)"
    exit 1
}

[ $# -eq 0 ] && usage

setup_ldflags() {
    LIB_DIR="$HOME/lib"
    if [ -d "$LIB_DIR" ] && [ -f "$LIB_DIR/libXxf86vm.so" ]; then
        export CGO_LDFLAGS="-L$LIB_DIR"
    fi
}

cmd_build() {
    setup_ldflags
    go build -o arbok ./cmd/arbok
    echo "Build successful: ./arbok"
}

cmd_test() {
    go test ./internal/terminal/ -v -count=1
}

cmd_vet() {
    go vet ./...
}

cmd_fmt() {
    if [ -n "$(gofmt -l .)" ]; then
        echo "Formatting issues:"
        gofmt -l .
        exit 1
    fi
}

cmd_lint() {
    cmd_vet
    cmd_fmt
}

cmd_all() {
    cmd_lint
    cmd_test
}

case "$1" in
    build) cmd_build ;;
    test)  cmd_test ;;
    vet)   cmd_vet ;;
    fmt)   cmd_fmt ;;
    lint)  cmd_lint ;;
    all)   cmd_all ;;
    *)     usage ;;
esac
