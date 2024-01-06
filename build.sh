#!/bin/bash

clean() {
  rm -rf dist/
}

echo "cleaning..."
clean

echo "go mod tidy"
go mod tidy

mkdir dist

echo "building to dist/btw.exe"
CGO_ENABLED=1 go build -ldflags -H=windowsgui -o dist/btw.exe