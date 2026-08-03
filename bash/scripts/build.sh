#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

source "${SCRIPT_DIR}/lib/lifecycle/build_runtime.sh"
source "${SCRIPT_DIR}/lib/lifecycle/lock.sh"

main() {
    local mode="${1:-}"
    local source="${2:-}"

    case "$mode" in
        modelist)
            modelist
            ;;
        build)
            if [[ -f "$source" && "$source" == *.yaml ]]
            then
                build_recipe "$source"
            elif [[ -d "$source" ]]
            then
                build_archive "$source"
            else
                build_recipe "$source"
            fi
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

modelist() {
    echo build
}

build_recipe() {
    local recipe_name="${1:-}"
    recipe_load "${recipe_name}"
    local recipe_name="$(recipe_name)"
    local dist_name="$(dist_build_name "$recipe_name")"
    dist_create "$dist_name" "$(recipe_file)" "$(recipe_platform)"

    create_platform
    create_profiles
    kill_platform

    setting_platform
    setting_profiles

    build_profiles
    platform_print_summary
    lock_after_mutation "$dist_name"
    lock_pool_update

    echo
    echo "[done] $dist_name"
}

build_archive() {
    local arc_dir="${1:-}"
    arc_load "${arc_dir}"
    lock_load "${arc_dir}/lock"
    recipe_load "$(arc_recipe)"
    local recipe_name="$(recipe_name)"
    local dist_name="$(dist_build_name "$recipe_name")"
    dist_create "$dist_name" "$(recipe_file)" "$(recipe_platform)"

    create_platform
    create_profiles
    kill_platform

    setting_platform "archive"
    setting_profiles "archive"

    build_profiles   "archive"
    platform_print_summary
    lock_after_mutation "$dist_name" "${arc_dir}/lock"

    echo
    echo "[done] $dist_name"
}

main "$@"
