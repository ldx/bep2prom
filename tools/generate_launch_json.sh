#!/usr/bin/env bash

set -euo pipefail

bazel_flags="--strip=never -c dbg"

workspace_root=$(bazel info workspace)

build_targets=$(bazel query 'kind(go_binary, //...) except attr(generator_name, image, //...)' --output label | sort)
test_targets=$(bazel query 'kind(go_test, //...)' --output label | sort)

tasks=()
configurations=()
for target in "${build_targets[@]}" "${test_targets[@]}"; do
    output=$(bazel cquery ${bazel_flags} --output=files "$target")
    echo "Adding ${target}, output: ${output}"
    configurations+=(
        "{
            \"name\": \"Launch ${target}\",
            \"type\": \"go\",
            \"request\": \"launch\",
            \"mode\": \"exec\",
            \"program\": \"\${workspaceRoot}/${output}\",
            \"env\": {},
            \"args\": [],
            \"preLaunchTask\": \"Build ${target}\",
            \"cwd\": \"\${workspaceRoot}\",
            \"showLog\": true,
            \"substitutePath\": [
                { \"from\": \"\${workspaceFolder}\", \"to\": \"\" }
            ],
            \"trace\": \"verbose\"
        }"
    )
    group="build"
    if [[ "${target}" == *"_test" ]]; then
        group="test"
    fi
    tasks+=(
        "{
            \"label\": \"Build ${target}\",
            \"type\": \"shell\",
            \"command\": \"bazel build ${bazel_flags} ${target}\",
            \"group\": \"${group}\",
            \"problemMatcher\": []
        }"
    )
done

echo "${tasks[@]}" | jq --indent 4 -s '{"version": "2.0.0", "tasks": .}' > "${workspace_root}/.vscode/tasks.json"

echo "${configurations[@]}" | jq --indent 4 -s '{"version": "0.2.0", "configurations": .}' > "${workspace_root}/.vscode/launch.json"
