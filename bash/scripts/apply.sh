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
        apply)
            if [[ -f "$source" && "$source" == *.yaml ]]
            then
                apply_recipe "$source"
            elif [[ -d "$source" && -f "${source}/recipe.yaml" && -f "${source}/settings.jsonc" ]]
            then
                apply_lock "$source"
            elif [[ -d "$source" ]]
            then
                apply_archive "$source"
            else
                apply_recipe "$source"
            fi
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

modelist() {
    echo apply
}

apply_recipe() {
    local recipe_file="${1:-}"
    recipe_load "${recipe_file}"
    local recipe_name="$(recipe_name)"
    dist_load "$recipe_name"

    [[ -f "${DIST_DIR}/run.sh" ]] && dist_create_run "$(recipe_platform)"
    [[ -f "${DIST_DIR}/$recipe_name" ]] && dist_create_exec "$recipe_name"

    create_profiles
    kill_platform

    setting_platform
    setting_profiles

    build_profiles
    dist_create_meta "${RECIPE_FILE}"
    lock_after_mutation "$(dist_name)"
    lock_pool_update

    echo
    echo "[done] apply $(recipe_name)"
}

apply_archive() {
    local arc_dir="${1:-}"
    arc_load "${arc_dir}"
    lock_load "${arc_dir}/lock"
    recipe_load "$(arc_recipe)"
    local recipe_name="$(recipe_name)"
    dist_load "$recipe_name"

    [[ -f "${DIST_DIR}/run.sh" ]] && dist_create_run "$(recipe_platform)"
    [[ -f "${DIST_DIR}/$recipe_name" ]] && dist_create_exec "$recipe_name"

    create_profiles
    kill_platform

    setting_platform "archive"
    setting_profiles "archive"

    build_profiles   "archive"
    dist_create_meta "${RECIPE_FILE}"
    lock_after_mutation "$(dist_name)" "${arc_dir}/lock"

    echo
    echo "[done] apply $(recipe_name)"
}

apply_lock() {
    local lock_dir="${1:-}"
    lock_load "$lock_dir"
    recipe_load "$(lock_recipe_file)"
    local recipe_name="$(recipe_name)"
    dist_load "$recipe_name"

    [[ -f "${DIST_DIR}/run.sh" ]] && dist_create_run "$(recipe_platform)"
    [[ -f "${DIST_DIR}/$recipe_name" ]] && dist_create_exec "$recipe_name"

    create_profiles
    kill_platform

    setting_platform "lock"
    setting_profiles "lock"

    platform_extension_recover "${DIST_DIR}"
    build_profiles "lock"
    dist_create_meta "$(lock_recipe_file)"
    lock_after_mutation "$(dist_name)" "$lock_dir"
    lock_pool_update

    echo
    echo "[done] apply $(recipe_name)"
}

main "$@"
