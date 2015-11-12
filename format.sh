#!/usr/bin/env bash

find . -path ./Godeps -prune -o -name "*.go" -exec goimports -w {} \; -exec gofmt -s -w {} \;
