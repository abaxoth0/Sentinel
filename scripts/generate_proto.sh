#!/bin/bash

PROJECT_ROOT=$(git rev-parse --show-toplevel)

find $PROJECT_ROOT/packages/common/proto -name '*.proto' | while read proto; do
  protoc --proto_path=$PROJECT_ROOT \
       --go_out=$PROJECT_ROOT \
       --go_opt=module=github.com/StepanAnanin/Sentinel \
       $PROJECT_ROOT/packages/common/proto/*.proto
done

