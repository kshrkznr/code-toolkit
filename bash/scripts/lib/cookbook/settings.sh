#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/find.sh"
source "${SCRIPT_DIR}/lib/json.sh"
source "${SCRIPT_DIR}/lib/cookbook/find.sh"
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"

settings_os_file() {
    settings_files "$(path_ingredient)" "os" "$1"
}

settings_platform_file() {
    settings_files "$(path_ingredient)" "platform" "$1"
}

settings_runtime_file() {
    settings_variant_files "runtime" "$1"
}

settings_profile_file() {
    settings_variant_files "profile" "$1"
}

settings_extension_file() {
    settings_variant_files "extension" "$1"
}

settings_variant_files() {
    local layer="$1"
    local ingredient="$2"
    local os="$(recipe_os)"
    local platform="$(recipe_platform)"

    settings_files "$(path_ingredient)" "$layer" "$ingredient"
    [[ -n "$os" ]] && settings_files "$(path_ingredient)" "$layer" "$ingredient" "${os}."
    [[ -n "$platform" ]] && settings_files "$(path_ingredient)" "$layer" "$ingredient" "${platform}."
}

settings_files() {
    local root="$1"
    local layer="$2"
    local ingredient="$3"
    local suffix="${4:-}"

    while read -r settings
    do
        [[ -f "${settings}" ]]  && echo "${settings}" && return
    done < <(find_file_thin "${root}" "${layer}.${ingredient}.${suffix}settings.json*")

    while read -r settings
    do
        [[ -f "${settings}" ]]  && echo "${settings}" && return
    done < <(find_file_thin "${root}/${layer}" "${ingredient}.${suffix}settings.json*")

    while read -r settings
    do
        [[ -f "${settings}" ]]  && echo "${settings}" && return
    done < <(find_file_thin "${root}/${layer}/${ingredient}" "${suffix}settings.json*")
}
