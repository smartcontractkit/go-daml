#!/bin/bash

# This script generates all test data
godaml --dar ./test.dar --output . --go_package main
mv ./rental_0_1_0.go ./rental_0_1_0.go_gen

godaml --dar ./test_2_9_1.dar --output . --go_package main
mv ./test_1_0_0.go ./test_1_0_0.go_gen

godaml --dar ./all-kinds-of-1.0.0_lf.dar --output . --go_package codegen_test
cp ./all_kinds_of_1_0_0.go ../examples/codegen/all_kinds_of_1_0_0.go
mv ./all_kinds_of_1_0_0.go ./all_kinds_of_1_0_0.go_gen
