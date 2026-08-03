#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"

source "${SCRIPT_DIR}/lib/recipe_resolver.sh"
source "${SCRIPT_DIR}/lib/platform.sh"
source "${SCRIPT_DIR}/lib/vsix.sh"
source "${SCRIPT_DIR}/lib/json.sh"

create_platform() {
    platform_create "${DIST_DIR}"
}

create_profiles() {

    local profile

    while read -r profile
    do
        create_profile "${profile}"
    done < <(recipe_profiles)

}

create_profile() {
    local profile="$1"
    platform_create_profile "${DIST_DIR}" "${profile}"
}

setting_platform() {
    local install_type="${1:-recipe}"

    resolve_settings "${install_type}" \
    | json_read \
    | json_merge \
    > "${DIST_DIR}/.data/User/settings.json"
}

setting_profiles() {
    local install_type="${1:-recipe}"
    while read -r profile
    do
        setting_profile "${profile}" "$install_type"
    done < <(recipe_profiles)

}

setting_profile() {
    local profile="$1"
    local install_type="${2:-recipe}"
    local settings_default keybindings_default tasks_default mcp_default snippets_default

    settings_default="$(recipe_profile_default_flag "$profile" settings)" || return 1
    keybindings_default="$(recipe_profile_default_flag "$profile" keybindings)" || return 1
    tasks_default="$(recipe_profile_default_flag "$profile" tasks)" || return 1
    mcp_default="$(recipe_profile_default_flag "$profile" mcp)" || return 1
    snippets_default="$(recipe_profile_default_flag "$profile" snippets)" || return 1

    platform_profile_contents "${DIST_DIR}" "${profile}" \
        "$settings_default" \
        "$keybindings_default" \
        "$tasks_default" \
        "$mcp_default" \
        "$snippets_default"

    [[ "$settings_default" == true ]] || setting_profile_settings "$profile" "$install_type"
}

setting_profile_settings() {
    local profile="$1"
    local install_type="${2:-recipe}"
    local settings_file
    settings_file="$(platform_profile_settings_file "${DIST_DIR}" "$profile")"

    mkdir -p "$(dirname "$settings_file")"
    resolve_profile_settings "$profile" "$install_type" \
        | json_read \
        | json_merge \
        > "$settings_file"
}

kill_platform() {
    platform_close_dist "$(dist_name)"
}

build_profiles() {
    local install_type="${1:-recipe}"
    build_runtime_extensions "${install_type}"

    while read -r profile
    do
        build_profile "${profile}" "${install_type}"
    done < <(recipe_profiles)

}

build_runtime_extensions() {
    local install_type="${1:-recipe}"
    local update_extension=true
    local extension_mode
    [[ "$install_type" == archive ]] && update_extension=false
    extension_mode="$(recipe_default_profile_extensions)"

    case "$extension_mode" in
    clean)
        echo "[runtime build] extensions: clean"
        platform_uninstall_extensions "${DIST_DIR}" "" "$update_extension" < /dev/null
        return
        ;;
    runtime)
        ;;
    *)
        echo "invalid default-profile.extensions: $extension_mode (expected: runtime, clean)" >&2
        return 1
        ;;
    esac

    echo "[runtime build] extensions"
    build_runtime_extension_ids "${install_type}" \
        | resolve_pool_extensions "${install_type}" \
        | platform_install_extensions "${DIST_DIR}" "" "$update_extension"

    build_runtime_extension_ids "${install_type}" \
        | platform_uninstall_extensions "${DIST_DIR}" "" "$update_extension"
}

build_profile() {
    local profile="$1"
    local install_type="${2:-}"
    local update_extension=true
    [[ "$install_type" == archive ]] && update_extension=false

    echo "[profile build] ${profile}"

    build_profile_extension_ids "${profile}" "${install_type}" \
        | resolve_pool_extensions "${install_type}" \
        | platform_install_extensions "${DIST_DIR}" "${profile}" "$update_extension"

    build_profile_extension_ids "${profile}" "${install_type}" \
        | platform_uninstall_extensions "${DIST_DIR}" "${profile}" "$update_extension"

    echo "[profile build finished] ${profile}"
}

build_runtime_extension_ids() {
    local install_type="${1:-recipe}"

    resolve_runtime_extensions "$install_type" \
        | extension_read "$install_type" \
        | extension_merge
}

build_profile_extension_ids() {
    local profile="$1"
    local install_type="${2:-recipe}"

    resolve_profile_extensions "$profile" "$install_type" \
        | extension_read "$install_type" \
        | extension_merge
}

resolve_pool_extensions() {
    local install_type="${1:-recipe}"

    if [[ "$install_type" == archive ]]
    then
        cat
        return
    fi

    vsix_pool_resolve "$(recipe_platform)" "$(recipe_extension_marketplace)"
}
