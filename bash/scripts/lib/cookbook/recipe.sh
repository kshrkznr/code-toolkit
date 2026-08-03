#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
RECIPE_FILE=""

recipe_file() {
    echo "${RECIPE_FILE}"
}

recipe_load() {
    local recipe_name="${1:-}"

    if [[ -z "$recipe_name" ]]
    then
        recipe_name="$(recipe_list | fzf)"
    fi

    [[ -z "$recipe_name" ]] && return 1

    if [[ -f "$recipe_name" && "$recipe_name" == *.yaml ]]
    then
        recipe_file_set "${recipe_name}"
    else
        recipe_file_set "$(path_recipe)/${recipe_name}"
    fi
}

recipe_file_set() {
    RECIPE_FILE="${1:-}"

    if [[ ! -f "$RECIPE_FILE" ]]
    then
        echo "recipe not found: $RECIPE_FILE" >&2
        return 1
    fi
    echo "recipe : $(recipe_name)"
}

recipe_list() {
    find_recipe | xargs -n1 basename
}

recipe_dir() {
    echo "$(path_recipe)"
}

recipe_name() {
    yq -r '.name // ""' "$RECIPE_FILE"
}

recipe_get_list() {
    local section="$1"

    yq -r ".${section}[]?" "$RECIPE_FILE"
}

recipe_profiles() {
    recipe_get_list profile
}

recipe_runtimes() {
    recipe_get_list runtime
}

recipe_platform() {
    yq -r '.platform // ""' "$RECIPE_FILE"
}

recipe_os() {
    yq -r '.os // ""' "$RECIPE_FILE"
}

recipe_extension_marketplace() {
    yq -r '.config."dist-strategy"."extension-marketplace" // true' "$RECIPE_FILE"
}

recipe_lock_mode() {
    yq -r '.config."dist-strategy"."lock-mode" // "refresh"' "$RECIPE_FILE"
}

recipe_default_profile_extensions() {
    yq -r '.config."dist-strategy"."default-profile".extensions // "runtime"' "$RECIPE_FILE"
}

recipe_profile_default_flag() {
    local profile="$1"
    local content="$2"
    local strategy
    strategy="$(PROFILE="$profile" CONTENT="$content" \
        yq -r '.config."profile-strategy"[strenv(PROFILE)][strenv(CONTENT)] // "default"' \
        "$RECIPE_FILE")"

    case "$strategy" in
        default)
            echo true
            ;;
        profile)
            echo false
            ;;
        *)
            echo "invalid profile-strategy.${profile}.${content}: $strategy (expected: default, profile)" >&2
            return 1
            ;;
    esac
}

recipe_exists_profile() {
    recipe_exists profile "$1"
}

recipe_exists_runtime() {
    recipe_exists runtime "$1"
}

recipe_exists() {
    local section="$1"
    local target="$2"

    recipe_get_list "$section" | grep -qx "$target"
}

recipe_validate() {

    [[ -n "$(recipe_name)" ]] || {
        echo "recipe.name missing" >&2
        return 1
    }

    recipe_profiles >/dev/null || return 1

    return 0
}

recipe_dump() {

    echo "name     : $(recipe_name)"

    echo "os"
    recipe_os | sed 's/^/  - /'

    echo "platform"
    recipe_platform | sed 's/^/  - /'

    echo "runtime"
    recipe_runtimes | sed 's/^/  - /'

    echo "profile"
    recipe_profiles | sed 's/^/  - /'
}
