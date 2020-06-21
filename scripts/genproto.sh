#!/usr/bin/env bash
#
# Generate all protobuf bindings.
# Run from repository root.
set -e
set -u

PROTOC_BIN=${PROTOC_BIN:-protoc}
GOIMPORTS_BIN=${GOIMPORTS_BIN:-goimports}
PROTOC_GEN_GO_BIN=${PROTOC_GEN_GO_BIN:-protoc-gen-go}
PROTOC_GEN_GO_GRPC_BIN=${PROTOC_GEN_GO_GRPC_BIN:-protoc-gen-go-grpc}
PROTOC_GEN_GOGOFAST_BIN=${PROTOC_GEN_GOGOFAST_BIN:-protoc-gen-gogofast}

if ! [[ "$0" =~ "scripts/genproto.sh" ]]; then
	echo "must be run from repository root"
	exit 255
fi

OLDPATH=${PATH}

mkdir -p /tmp/protobin/
cp ${PROTOC_GEN_GO_BIN} /tmp/protobin/protoc-gen-go
cp ${PROTOC_GEN_GO_GRPC_BIN} /tmp/protobin/protoc-gen-go-grpc
PATH=${OLDPATH}:/tmp/protobin

DIRS="grpctesting/testpb"
echo "generating protobuf code"
for dir in ${DIRS}; do
	pushd ${dir}
		${PROTOC_BIN} --go_out=. --go-grpc_out=.\
      -I=. \
			*.proto

			${GOIMPORTS_BIN} -w *.pb.go
	popd
done
