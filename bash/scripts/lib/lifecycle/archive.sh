#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/lifecycle/find.sh"
source "${SCRIPT_DIR}/lib/lifecycle/lock.sh"
ARC_DIR=""

arc_create() {
    local dist_name="${1:-}"
    local dist_dir="${2:-}"
    ARC_DIR="$(path_archive)/${dist_name}"

    lock_ensure_lock "$dist_name"

    rm -rf "${ARC_DIR}"
    mkdir -p "${ARC_DIR}"
    cp -r "${dist_dir}/.lock" "${ARC_DIR}/lock"
    cp "${dist_dir}/.meta/recipe.yaml" "${ARC_DIR}/lock/recipe.yaml"
    cp -r "${dist_dir}/.meta" "${ARC_DIR}/.dist.meta"
    mkdir -p "${ARC_DIR}/vsix"
    arc_create_meta "$dist_dir"
}

arc_create_meta() {
    local dist_dir="${1:-}"
    mkdir -p "${ARC_DIR}/.arc.meta"
    cat > "${ARC_DIR}/.arc.meta/arc.env" <<EOF
DIST=$(basename "${dist_dir}")
ARC_TIME=$(date '+%Y-%m-%d %H:%M:%S')
EOF
}

arc_load() {
    local arc_name="${1:-}"

    if [[ -z "$arc_name" ]]
    then
        arc_name="$(arc_list | fzf)"
    fi

    [[ -z "$arc_name" ]] && return 1

    if [[ -d "${arc_name}" ]]
    then
        arc_dir_set "${arc_name}"
    else
        arc_dir_set "$(path_archive)/${arc_name}"
    fi
}

arc_dir_set() {
    ARC_DIR="${1:-}"
    if [[ ! -d "$ARC_DIR" ]]
    then
        echo "dist not found: $ARC_DIR" >&2
        return 1
    fi
    ARC_DIR="$(cd "$ARC_DIR" && pwd)"
    echo "arc : $ARC_DIR"
}

arc_name() {
    basename "${ARC_DIR}"
}

arc_recipe() {
    echo "${ARC_DIR}/lock/recipe.yaml"
}

arc_extensions_lock_file() {
    local profile="$1"
    echo "${ARC_DIR}/lock/${profile}.extensions.lock"
}

arc_profile_settings_file() {
    local profile="$1"
    echo "${ARC_DIR}/lock/${profile}.settings.jsonc"
}

arc_list() {
    find_archive | xargs -n1 basename
}
