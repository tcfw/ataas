#!/bin/bash

if [ -z `which protoc` ]; then
        echo "Error: protoc was not found in the \$PATH variable or is not installed";
        exit 1;
fi

files=$(find ./protos -type f -name "*.proto" | xargs basename -s ".proto")

mkdir -p api/swagger

for service in $files
do
        echo -e "\xe2\x84\xb9\xef\xb8\x8f  Generating '${service}' API...";
        mkdir -p api/pb/${service}

        #GRPC
        protoc --proto_path=protos/ \
                --gogofaster_out=\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
paths=source_relative:api/pb/${service} \
                --swagger_out=logtostderr=true:api/swagger \
                --grpc-gateway_out=logtostderr=true,paths=source_relative:api/pb/${service} \
                -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
                -I=$GOPATH/src/github.com/gogo/protobuf/gogoproto \
                -I=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
                --go-grpc_out=\
Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
paths=source_relative:api/pb/${service} \
                ${service}.proto

done