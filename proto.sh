#!/bin/bash

INPUT_FOLDER="../proto-definitions"
OUTPUT_FOLDER="./protobuf"

# grep the go module from the go mod
GO_MOD=$(grep module go.mod | sed 's/^module //')

# build_microservice builds a microservice given its name
# $1 name of the microservice
build_microservice() {

    # create a folder for each microservice
    mkdir -p "$OUTPUT_FOLDER"/"$1"

    # genereate the goimport
    GO_IMPORT="$GO_MOD"/"$OUTPUT_FOLDER"/"$1"

    # List all proto files
    protos="$INPUT_FOLDER/$1/*.proto"

    imports=""
    for proto in $protos; do
        base_proto=$(basename "$proto")
        # generate the go_opt and grpc_server imports for each one
        imports+="\
            --go_opt=M$1/$base_proto=$GO_IMPORT \
            --go-grpc_opt=M$1/$base_proto=$GO_IMPORT \
        "
        # print the proto_definition and the proto-file
        echo "[$1]: $base_proto"
    done

    # actually want to split the content of those variable based on spaces
    # shellcheck disable=SC2086
    protoc \
        --proto_path=./"$INPUT_FOLDER"/ \
        --go_out=./"$OUTPUT_FOLDER"/ \
        --go_opt=paths=source_relative \
        --go-grpc_out=./"$OUTPUT_FOLDER"/ \
        $imports \
        --go-grpc_opt=paths=source_relative \
        $protos \
        --experimental_allow_proto3_optional
}

# Remove existing generated files
rm -rf "$OUTPUT_FOLDER"

# Build each microservice
for m in "$INPUT_FOLDER"/*/; do
    build_microservice "$(basename "$m")" &
done

wait