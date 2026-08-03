#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/find.sh"
source "${SCRIPT_DIR}/lib/cookbook/find.sh"

extension_runtime_file() {
    extension_files "$(path_ingredient)" "runtime" "$1"
}

extension_profile_file() {
    extension_files "$(path_ingredient)" "profile" "$1"
}

extension_files() {
    local root="$1"
    local layer="$2"
    local ingredient="$3"
    while read -r extension

    do
        [[ -f "${extension}" ]]  && echo "${extension}" && return
    done < <(find_file_thin "${root}" "${layer}.${ingredient}.extensions")

    while read -r extension
    do
        [[ -f "${extension}" ]]  && echo "${extension}" && return
    done < <(find_file_thin "${root}/${layer}" "${ingredient}.extensions")

    find_extension "${root}/${layer}/${ingredient}"
}

extension_merge() {
    sort -u
}

extension_read() {
    local install_type="$1"
    while read -r file; do
        case "$install_type" in
            archive)
                echo "${file}"
                ;;
            lock)
                echo "${file}"
                ;;
            *)
                cat "${file}"
                ;;
        esac
    done
}
