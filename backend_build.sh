#!/bin/bash

set -ex

go test -timeout 1m -failfast -tags=integration -run "${RUN_TESTS:-.+}" -v ./pkg/ &&
GOOS=linux GOARCH=amd64 go build -o ./dist/bilibili-altinity-clickhouse-plugin_linux_amd64 ./pkg/ &&
GOOS=linux GOARCH=arm64 go build -o ./dist/bilibili-altinity-clickhouse-plugin_linux_arm64 ./pkg/ &&
GOOS=windows GOARCH=amd64 go build -o ./dist/bilibili-altinity-clickhouse-plugin_windows_amd64.exe ./pkg/ &&
GOOS=darwin GOARCH=amd64 go build -o ./dist/bilibili-altinity-clickhouse-plugin_darwin_amd64 ./pkg/ &&
GOOS=darwin GOARCH=arm64 go build -o ./dist/bilibili-altinity-clickhouse-plugin_darwin_arm64 ./pkg/ &&
chmod +x ./dist/bilibili-altinity-clickhouse-plugin*
