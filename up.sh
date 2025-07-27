#!/usr/bin/env bash

echo "Uploading the web content"

go generate ./cmd/sim

patterns=(*.html *.css *.js *.wasm)
for p in "${patterns[@]}"; do
  gcloud storage cp ./cmd/sim/"$p" gs://rmazur-io-fahivets/fahivets-sim/ --recursive
done

echo "DONE!"
