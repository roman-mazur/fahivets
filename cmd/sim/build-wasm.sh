#!/usr/bin/env bash

cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
mkdir -p ./internal/ui/progs/

progs_src=../../testdata/progs
progs_dst=./internal/ui/progs
cp -r $progs_src/*.rom $progs_dst
cp $progs_src/*.rks $progs_dst

GOOS=js GOARCH=wasm go build -o main.wasm ./internal/ui
