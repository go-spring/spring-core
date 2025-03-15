#!/usr/bin/env bash

export GOCOVERDIR=cov
rm -rf ./cov/*
go run -race -cover -covermode=atomic main.go
go tool covdata textfmt -i=cov -o ./cov/cover.txt
go tool cover -html=./cov/cover.txt -o ./cov/cover.html