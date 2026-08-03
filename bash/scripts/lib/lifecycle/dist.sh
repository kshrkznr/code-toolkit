#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/os.sh"
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"
source "${SCRIPT_DIR}/lib/lifecycle/find.sh"
DIST_DIR=""

dist_load() {
    local dist_name="${1:-}"

    if [[ -z "$dist_name" ]]
    then
        dist_name="$(dist_list | fzf)"
    fi

    [[ -z "$dist_name" ]] && return 1

    if [[ -d "${dist_name}" ]]
    then
        dist_dir_set "${dist_name}"
    else
        dist_dir_set "$(path_dist)/${dist_name}"
    fi

}

dist_dir_set() {
    DIST_DIR="${1:-}"
    if [[ ! -d "$DIST_DIR" ]]
    then
        echo "dist not found: $DIST_DIR" >&2
        return 1
    fi
    DIST_DIR="$(cd "$DIST_DIR" && pwd)"
    echo "dist : $DIST_DIR"
}

dist_create() {
    local dist_name="${1:-}"
    local recipe_file="${2:-}"
    local recipe_platform="${3:-}"
    DIST_DIR="$(path_dist)/${dist_name}"
    rm -rf "${DIST_DIR}"
    mkdir -p "${DIST_DIR}"
    dist_load "$dist_name"

    mkdir -p "${DIST_DIR}/.data"
    mkdir -p "${DIST_DIR}/.ext"
    mkdir -p "${DIST_DIR}/.meta"
    dist_create_run "$recipe_platform"
    dist_create_exec "$dist_name"
    dist_create_meta "$recipe_file"
}

dist_build_name() {
    local base_name="${1:-}"
    local candidate="$base_name"
    local suffix=1

    [[ -n "$base_name" ]] || {
        echo "dist name missing" >&2
        return 1
    }

    while [[ -e "$(path_dist)/${candidate}" || -L "$(path_dist)/${candidate}" ]]
    do
        candidate="${base_name}.${suffix}"
        ((suffix += 1))
    done

    printf '%s\n' "$candidate"
}

dist_create_meta() {
    local recipe_file="${1:-}"
    [[ -n "$recipe_file" && -f "$recipe_file" ]] || {
        echo "recipe file missing for dist metadata" >&2
        return 1
    }

    cp "${recipe_file}" "$(dist_recipe)"

    cat > "${DIST_DIR}/.meta/build.env" <<EOF
RECIPE=$(basename "${recipe_file}")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
EOF
}

dist_create_run() {
    os_run_sh "${1:-}" > "${DIST_DIR}/run.sh"
    chmod +x "${DIST_DIR}/run.sh"
}

dist_create_exec() {
    local script_name="$1"
    local exec_file="$(os_exec_file "$script_name")"
    chmod +x "${exec_file}"
}

dist_name() {
    local dist_dir=${1:-$DIST_DIR}
    basename "${dist_dir}"
}

dist_recipe() {
    echo "${DIST_DIR}/.meta/recipe.yaml"
}

dist_list() {
    find_dist | xargs -n1 basename
}

dist_is_current() {
    local dist_dir="${1:-$DIST_DIR}"
    local platform current_name
    platform="$(recipe_load "${dist_dir}/.meta/recipe.yaml" >/dev/null; recipe_platform)"
    [[ -n "$platform" && "$platform" != "null" ]] || return 1
    current_name=$("$(path_script)/codevenv.sh" "current" "$platform")
    if [[ $(dist_name "$dist_dir") == "$current_name" ]]
    then
        return 0
    fi
    return 1
}
