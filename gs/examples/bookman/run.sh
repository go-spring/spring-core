#!/usr/bin/env bash

export GOCOVERDIR=.cover
rm -rf ./.cover/*
go run -race -cover -covermode=atomic main.go
go tool covdata textfmt -i=.cover -o ./.cover/cover.txt
go tool cover -html=./.cover/cover.txt -o ./.cover/cover.html