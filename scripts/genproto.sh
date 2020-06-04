#!/usr/bin/env bash
#
# Generate all protobuf bindings.
# Run from repository root.
set -e
set -u

PROTOC_BIN=${PROTOC_BIN:-protoc}
GOIMPORTS_BIN=${GOIMPORTS_BIN:-goimports}

if ! [[ "$0" =~ "scripts/genproto.sh" ]]; then
	echo "must be run from repository root"
	exit 255
fi

echo "installing gogofast"
GO111MODULE=on go install "github.com/gogo/protobuf/protoc-gen-gogofast"

GOGOPROTO_ROOT="$(GO111MODULE=on go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)"
GOGOPROTO_PATH="${GOGOPROTO_ROOT}:${GOGOPROTO_ROOT}/protobuf"

DIRS="grpctesting/testpb grpctesting/gogotestpb"
echo "generating code"
for dir in ${DIRS}; do
	pushd ${dir}
		${PROTOC_BIN} --gogofast_out=plugins=grpc:. -I=. \
			-I="${GOGOPROTO_PATH}" \
			*.proto
	popd
done
