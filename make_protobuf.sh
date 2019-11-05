#!/bin/bash
# Generates the go code for the protobuf specifications.
# Make sure that `protoc-gen-go` is discoverable from your $PATH variable.

protoc --go_out=. config/structs.proto
protoc --go_out=. sphinx/sphinx_structs.proto

# Temporary workaround until the remaining code is fixed to use 
# proper key: value identifiers for the protobuf structs #FIXME
sed -i '/XXX_NoUnkeyedLiteral/d' config/structs.pb.go
sed -i '/XXX_unrecognized/d' config/structs.pb.go
sed -i '/XXX_sizecache/d' config/structs.pb.go
sed -i '/XXX_NoUnkeyedLiteral/d' sphinx/sphinx_structs.pb.go
sed -i '/XXX_unrecognized/d' sphinx/sphinx_structs.pb.go
sed -i '/XXX_sizecache/d' sphinx/sphinx_structs.pb.go