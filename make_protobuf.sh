#!/bin/bash
# Generates the go code for the protobuf specifications.
# Make sure that `protoc-gen-go` is discoverable from your $PATH variable.

protoc --go_out=. config/*.proto
protoc --go_out=. sphinx/*.proto
