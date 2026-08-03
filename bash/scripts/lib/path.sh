#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/os.sh"
CTK_HOME="$(cd "${SCRIPT_DIR}/../.." && pwd)"

path_home() {
    echo "${CTK_HOME}"
}

path_dist() {
    echo "${CTK_HOME}/dist"
}

path_archive() {
    echo "${CTK_HOME}/archive"
}

path_vsix() {
    echo "${CTK_HOME}/.vsix"
}

path_inspect() {
    echo "${CTK_HOME}/cookbook/inspect"
}

path_draft() {
    echo "${CTK_HOME}/cookbook/draft"
}

path_old() {
    echo "${CTK_HOME}/.old"
}

path_cookbook() {
    echo "${CTK_HOME}/cookbook"
}

path_recipe() {
    echo "${CTK_HOME}/cookbook/recipe"
}

path_ingredient() {
    echo "${CTK_HOME}/cookbook/ingredient"
}

path_os() {
    echo "${CTK_HOME}/cookbook/ingredient/os"
}

path_platform() {
    echo "${CTK_HOME}/cookbook/ingredient/platform"
}

path_runtime() {
    echo "${CTK_HOME}/cookbook/ingredient/runtime"
}

path_profile() {
    echo "${CTK_HOME}/cookbook/ingredient/profile"
}

path_extension() {
    echo "${CTK_HOME}/cookbook/ingredient/extension"
}

path_script() {
    echo "${CTK_HOME}/scripts"
}

path_platform_strage(){
    local home="${1:-}"
    echo "${home}/User/globalStorage/storage.json"
}

path_platform_profile(){
    local home="${1:-}"
    local location="${2:-}"
    echo "${home}/User/profiles/${location}"
}
