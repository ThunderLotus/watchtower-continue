#!/bin/bash

cd $(git rev-parse --show-toplevel)

WASM_EXEC_JS="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "$WASM_EXEC_JS" ]; then
  WASM_EXEC_JS="$(go env GOROOT)/misc/wasm/wasm_exec.js"
fi
cp "$WASM_EXEC_JS" ./docs/assets/

GOARCH=wasm GOOS=js go build -o ./docs/assets/tplprev.wasm ./tplprev