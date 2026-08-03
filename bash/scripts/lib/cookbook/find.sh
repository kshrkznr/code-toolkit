#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/find.sh"

find_extension_roots() {
    path_ingredient
    find_runtime
    path_runtime
    find_runtime
    path_profile
    find_profile
}

find_settings_roots() {
    path_ingredient
    path_os
    find_os
    path_platform
    find_platform
    path_runtime
    find_runtime
    path_profile
    find_profile
}

find_recipe() {
    find_file_thin "$(path_recipe)" "*.yaml"
    find_file_thin "$(path_recipe)" "*.yml"

}

find_os() {
    find_dir_thin "$(path_os)"
}

find_platform() {
    find_dir_thin "$(path_platform)"
}

find_runtime() {
    find_dir_thin "$(path_runtime)"
}

find_profile() {
    find_dir_thin "$(path_profile)"
}

find_settings() {
    stdin_or_args "$@" | while read -r dir
    do
        find_file_thin "${dir}" "settings.json*"
        find_file_thin "${dir}" "*.settings.json*"
    done
}

find_extension() {
    stdin_or_args "$@" | while read -r dir
    do
        find_file_thin "${dir}" "extensions"
        find_file_thin "${dir}" "*.extensions"
    done
}

stdin_or_args() {
    if (($#)); then
        printf '%s\n' "$@"
    else
        cat
    fi
}
