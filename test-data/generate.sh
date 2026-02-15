#!/bin/bash

# This script generates all test data

godaml --dar ./all-kinds-of-1.0.0_lf.dar --output . --go_package codegen_test
cp ./all_kinds_of_1_0_0.go ../examples/codegen/all_kinds_of_1_0_0.go
mv ./all_kinds_of_1_0_0.go ./all_kinds_of_1_0_0.go_gen
