#!/usr/bin/env bash

part="${1:-TestAcc.*}"
TF_ACC=1 go test -v -run "${part}" -timeout 240m ./...