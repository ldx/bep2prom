#!/usr/bin/env bash

set -euo pipefail

workspace_root=$(bazel info workspace)

targets=$(bazel query 'kind(go_binary, //...) except attr(generator_name, image, //...)' --output label | sort)

tasks=()
configurations=()
for target in "${targets[@]}"; do
    output=$(bazel cquery --output=files "$target")
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
    tasks+=(
        "{
            \"label\": \"Build ${target}\",
            \"type\": \"shell\",
            \"command\": \"bazel build --strip=never -c dbg ${target}\",
            \"group\": \"build\",
            \"problemMatcher\": []
        }"
    )
done

echo "${tasks[@]}" | jq --indent 4 -s '{"version": "2.0.0", "tasks": .}' > "${workspace_root}/.vscode/tasks.json"

echo "${configurations[@]}" | jq --indent 4 -s '{"version": "0.2.0", "configurations": .}' > "${workspace_root}/.vscode/launch.json"
