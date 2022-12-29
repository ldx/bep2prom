#!/usr/bin/env bash

# https://github.com/bazelbuild/rules_go/wiki/Editor-setup 
exec bazelisk run -- @io_bazel_rules_go//go/tools/gopackagesdriver "${@}"
