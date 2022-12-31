#!/usr/bin/env bash

set -euo pipefail

bazel_flags="--strip=never -c dbg"

workspace_root=$(bazel info workspace)

# Generate debugging entries for Go binary and test targets.
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

#
# For now, we can hardcode the path to dlv in settings.json via bazel-bin
# instead of the logic below.
#
## Update .vscode/settings.json with paths to tools the Go extension needs.
#tools=(dlv)
#declare -A tool_paths
#for tool in "${tools[@]}"; do
#    bazel build "//:${tool}"
#    tool_path=$(bazel cquery --output=files "//:${tool}")
#    tool_paths["${tool}"]="${tool_path}"
#done
#
#settings="{}"
#settings_path="${workspace_root}/.vscode/settings.json"
#if [[ -f "${settings_path}" ]]; then
#    settings=$(cat "${settings_path}")
#fi
#
#for tool in "${!tool_paths[@]}"; do
#    tool_path="${tool_paths[${tool}]}"
#    echo "Adding ${tool} => ${tool_path} to settings.json"
#    settings=$(echo "${settings}" | jq --arg tool "${tool}" --arg tool_path "${tool_path}" '. + {"go.alternateTools": {($tool): $tool_path}}')
#done
#
#echo "${settings}" | jq --indent 4 '.' > "${settings_path}"
