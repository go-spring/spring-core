#!/usr/bin/env bash

export GOCOVERDIR=cover
rm -rf ./cover/*
go run -race -cover -covermode=atomic main.go
go tool covdata textfmt -i=cover -o cover.txt
go tool cover -html=cover.txt -o cover.html