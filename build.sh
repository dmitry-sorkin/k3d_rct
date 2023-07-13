#!/bin/bash

rm assets/wasm/rct_lib.wasm
GOOS=js GOARCH=wasm go build -o assets\wasm\rct_lib.wasm main.go