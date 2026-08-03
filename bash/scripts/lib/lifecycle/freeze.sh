#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/cookbook/find.sh"
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"
source "${SCRIPT_DIR}/lib/cookbook/settings.sh"
source "${SCRIPT_DIR}/lib/cookbook/extension.sh"
source "${SCRIPT_DIR}/lib/lifecycle/find.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/lifecycle/lock.sh"
source "${SCRIPT_DIR}/lib/recipe_resolver.sh"
source "${SCRIPT_DIR}/lib/json.sh"

FREEZE_WORKBENCH=""

freeze_workspace() {
    echo "${FREEZE_WORKBENCH:-$(path_draft)}"
}

freeze_prepare() {
    rm -rf "$(freeze_workspace)"
    mkdir -p "$(freeze_workspace)/ref" "$(freeze_workspace)/lock" "$(freeze_workspace)/recipe"
}

freeze_collect_lock() {
    local dist_name="$1"
    local ensure_lock="${2:-true}"

    [[ "$ensure_lock" == true ]] && lock_ensure_lock "$dist_name"
    [[ -d "${DIST_DIR}/.lock" ]] || {
        echo "lock not found: ${DIST_DIR}/.lock" >&2
        return 1
    }

    rm -rf "$(freeze_workspace)/lock"
    cp -r "${DIST_DIR}/.lock" "$(freeze_workspace)/lock"
    lock_unlock_extension_files "$(freeze_workspace)/lock"
    find_extension_lock "$(freeze_workspace)/lock" \
        | xargs -n1 rm
}

freeze_collect_recipe() {
    local home="$(path_ingredient)"
    while read -r profile
    do
        while read -r ingredient
        do
            local path="$(freeze_workspace)/recipe/${ingredient#$home/}"
            mkdir -p "$(dirname ${path})"
            [[ ! -f ${ingredient} ]] && touch "${path}" && continue
            cp "${ingredient}" "${path}"
        done < <(resolve_profile_extensions "${profile}" "recipe")
    done < <(recipe_profiles)

    while read -r ingredient
    do
        local path="$(freeze_workspace)/recipe/${ingredient#$home/}"
        mkdir -p "$(dirname ${path})"
        [[ ! -f ${ingredient} ]] && touch "${path}" && continue
            cp "${ingredient}" "$(freeze_workspace)/recipe/${ingredient#$home/}"
    done < <(resolve_settings)

    local settings_default
    while read -r profile
    do
        settings_default="$(recipe_profile_default_flag "$profile" settings)" || return 1
        [[ "$settings_default" == false ]] || continue
        while read -r ingredient
        do
            local path="$(freeze_workspace)/recipe/${ingredient#$home/}"
            mkdir -p "$(dirname "$path")"
            [[ ! -f "$ingredient" ]] && touch "$path" && continue
            cp "$ingredient" "$path"
        done < <(resolve_profile_settings_content "$profile")
    done < <(recipe_profiles)
}

freeze_collect_reference() {
    local default_extension_mode
    default_extension_mode="$(recipe_default_profile_extensions)"
    case "$default_extension_mode" in
    clean)
        : > "$(freeze_workspace)/ref/runtime.extensions"
        ;;
    runtime)
        resolve_runtime_extensions "recipe" \
            | extension_read "recipe" \
            | extension_merge \
            > "$(freeze_workspace)/ref/runtime.extensions"
        ;;
    *)
        echo "invalid default-profile.extensions: $default_extension_mode (expected: runtime, clean)" >&2
        return 1
        ;;
    esac

    local settings_default
    while read -r profile
    do
        resolve_profile_extensions "${profile}" "recipe"  \
            | xargs -n1 cat | extension_merge \
            > "$(freeze_workspace)/ref/${profile}.extensions"

        settings_default="$(recipe_profile_default_flag "$profile" settings)" || return 1
        [[ "$settings_default" == false ]] || continue
        resolve_profile_settings "$profile" recipe \
            | json_read \
            | json_merge \
            > "$(freeze_workspace)/ref/${profile}.settings.jsonc"
    done < <(recipe_profiles)
    resolve_settings | json_read | json_merge > "$(freeze_workspace)/ref/settings.jsonc"
}

freeze_generate_extensions() {
    local dist_name="$1"
    local recipe_name="$2"
    local ext_draft="$(freeze_workspace)/extensions.draft"
    local home="$(path_ingredient)"

    echo "[do] draft : extensions"

    printf "# *** Inventory ***\n" > "${ext_draft}"
    find_extension_roots | find_extension \
    | while read -r path
    do
        echo "## ${path#$home/}"
    done >> "${ext_draft}"

    printf "\n\n# *** Difference ***\n">> "${ext_draft}"
    local difference=false

    while read -r path; do
        local extensions_file=$(basename "$path")
        local ref="$(freeze_workspace)/ref/${extensions_file}"
        local lock="$(freeze_workspace)/lock/${extensions_file}"
        set +e
        diff -q "${ref}" "${lock}" > /dev/null
        res=$?
        set -e
        if [[ $res -eq 0 ]] ; then
            echo "[no diff] $(basename "$path")"
        else
            difference=true
            echo "[diff] $(basename "$path")"
            set +eo > /dev/null
            {
                if [[ "$extensions_file" == "runtime.extensions" ]]
                then
                    echo "## runtime.draft.extensions"
                else
                    echo "## profile/${extensions_file}"
                fi
                printf "\n%s\n" '```diff'
                diff -u0bBN --suppress-common-lines \
                    --label "recipe_ref ($recipe_name)" --label "runtime    ($dist_name)"\
                    "${ref}" "${lock}"
                printf "%s\n\n" '```'
                } >> "${ext_draft}"
            set -eo > /dev/null
        fi
    done < <(freeze_lock_extension_files)

    ! "$difference" && echo "[No differences] extensions" >> "${ext_draft}"
    echo "[done] draft : extensions"

}

freeze_lock_extension_files() {
    local lock_dir="$(freeze_workspace)/lock"

    [[ -f "${lock_dir}/runtime.extensions" ]] && echo "${lock_dir}/runtime.extensions"
    find "$lock_dir" -maxdepth 1 -type f -name '*.extensions' \
        ! -name 'runtime.extensions' -print | sort
}

freeze_generate_settings() {
    local dist_name="$1"
    local recipe_name="$2"

    local setting_draft="$(freeze_workspace)/settings.draft"
    local home="$(path_ingredient)"
    echo "[do] draft : settings"
    printf "# *** Inventory ***\n" > "${setting_draft}"
    find_settings_roots | find_settings \
    | while read -r path; do
        echo "## ${path#$home/}"
    done >> "$setting_draft"

    printf "\n\n# *** Difference ***\n">> "${setting_draft}"

    local difference=false

    while IFS= read -r lock_file
    do
        local settings_file="$(basename "$lock_file")"
        local ref_file="$(freeze_workspace)/ref/${settings_file}"
        local header
        header="$(freeze_settings_draft_header "$settings_file")" || continue

        if freeze_generate_settings_difference "$ref_file" "$lock_file" "$header" \
            "$recipe_name" "$dist_name" >> "${setting_draft}"
        then
            difference=true
            echo "[diff] ${settings_file}"
        else
            echo "[no diff] ${settings_file}"
        fi
    done < <(freeze_lock_settings_files)

    if [[ "$difference" == false ]]
    then
        echo "[No differences] settings" >> "${setting_draft}"
    fi
    echo "[done] draft : settings"
}

freeze_lock_settings_files() {
    local lock_dir="$(freeze_workspace)/lock"

    [[ -f "${lock_dir}/settings.jsonc" ]] && echo "${lock_dir}/settings.jsonc"
    find "$lock_dir" -maxdepth 1 -type f -name '*.settings.jsonc' -print | sort
}

freeze_settings_draft_header() {
    local settings_file="$1"

    if [[ "$settings_file" == settings.jsonc ]]
    then
        echo 'runtime.draft.settings.json'
        return
    fi
    [[ "$settings_file" == *.settings.jsonc ]] || return 1
    echo "profile/${settings_file%.settings.jsonc}.settings.json"
}

freeze_generate_settings_difference() {
    local ref_file="$1"
    local lock_file="$2"
    local header="$3"
    local recipe_name="$4"
    local dist_name="$5"
    local ref_gron lock_gron ref lock

    ref_gron="$(mktemp)"
    lock_gron="$(mktemp)"
    ref="$(mktemp)"
    lock="$(mktemp)"
    json_read "$ref_file" | json_gron > "$ref_gron"
    json_read "$lock_file" | json_gron > "$lock_gron"
    json_gron_group "$ref_gron" > "$ref"
    json_gron_group "$lock_gron" > "$lock"

    if diff -q "$ref" "$lock" >/dev/null
    then
        rm -f "$ref_gron" "$lock_gron" "$ref" "$lock"
        return 1
    fi

    {
        echo "## ${header}"
        printf "\n%s\n" '```diff'
        diff -u0bBN --suppress-common-lines \
            --label "recipe_ref ($recipe_name)" --label "runtime    ($dist_name)" \
            "$ref" "$lock" || true
        printf "%s\n\n" '```'
    } | sed -E 's/^([+-])(\/\/ )/\2/' | sed -E 's/^[+-]$//'
    rm -f "$ref_gron" "$lock_gron" "$ref" "$lock"
}

freeze_generate_recipe(){
   cp "$(freeze_workspace)/lock/recipe.yaml" "$(freeze_workspace)/recipe.draft.yaml"
}

freeze_commit() {
    local draft="${1:-}"
    echo "[read] $draft"

    local ingredient_file=""
    local tmp=""
    while IFS= read -r line
    do
        line="${line%$'\r'}"
        case "$line" in
            "--- "*|"+++ "*|"@@ "*)
                continue
                ;;

            "## "*)
                freeze_commit_ingredient "$ingredient_file" "$tmp"
                ingredient_file="$(path_ingredient)/${line#\#\# }"
                backup_ingredient "$ingredient_file" "${line#\#\# }"
                tmp="$(mktemp)"
                freeze_prepare_ingredient "$ingredient_file" "$tmp"
                ;;

            "+"*)
                local ext="${line#+}"

                grep -qxF "$ext" "$tmp" \
                    || echo "$ext" >> "$tmp"
                ;;

            "-"*)
                local ext="${line#-}"

                grep -vxF "$ext" "$tmp" > "$tmp.new"
                mv "$tmp.new" "$tmp"
                ;;

            *)
                ;;
        esac
    done < "$draft"
    freeze_commit_ingredient "$ingredient_file" "$tmp"
    echo "[commit] $draft"
}

freeze_prepare_ingredient() {
    local ingredient_file="$1"
    local tmp="$2"

    [[ ! -f "$ingredient_file" ]] && return
    echo "[commit base] ${ingredient_file}"
    case "${ingredient_file##*.}" in
        json|jsonc)
            json_read "$ingredient_file" | json_gron > "$tmp"
            ;;
        *)
            cp "$ingredient_file" "$tmp"
            ;;
    esac
}

freeze_commit_ingredient() {
    local ingredient_file="$1"
    local tmp="$2"
    [[ -z "${tmp:-}" ]] && return
    [[ ! -s "$tmp" ]] && echo '{}' > "${ingredient_file}" && return

    mkdir -p "$(dirname ${ingredient_file})"
    echo "[commit to] ${ingredient_file}"
    case "${ingredient_file##*.}" in
        json|jsonc)
            cat "$tmp" | json_ungron > "${ingredient_file}"
            ;;
        extensions)
            cp "$tmp" "$ingredient_file"
            ;;
    esac
    rm -f "$tmp"
}

backup_ingredient() {
    local ingredient_file="$1"
    local ingredient_path="$2"
    local old_file="$(path_old)/${ingredient_path}"

    mkdir -p "$(dirname "$old_file")"
    [[ -f "${ingredient_file}" ]] && cp "$ingredient_file" "$old_file"
    :
    return
}

freeze_commit_recipe() {
    local recipe_file="${1:-}"
    echo "[read] $recipe_file"
    recipe_load "${recipe_file}"
    local name=$(recipe_name).$(recipe_os).yaml
    mkdir -p "$(path_recipe)"
    cp $recipe_file "$(path_recipe)/${name}"
    echo "[commit] $recipe_file"
}
