#!/usr/bin/env bash

cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
mkdir -p ./internal/ui/progs/
cp -r ../../testdata/progs/*.rom ./internal/ui/progs/

GOOS=js GOARCH=wasm go build -o main.wasm ./internal/ui
