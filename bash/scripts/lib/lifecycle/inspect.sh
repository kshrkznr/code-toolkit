#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/cookbook/find.sh"
source "${SCRIPT_DIR}/lib/cookbook/extension.sh"
source "${SCRIPT_DIR}/lib/recipe_resolver.sh"
source "${SCRIPT_DIR}/lib/json.sh"
source "${SCRIPT_DIR}/lib/lifecycle/freeze.sh"

inspect_prepare() {
    mkdir -p "$(path_inspect)"
}

inspect_view() {
    local source="${1:-}"
    local recipe_source dist_source

    [[ -n "$source" ]] || {
        echo "usage: inspect view <recipe.yaml | dist-dir | ingredient-dir>" >&2
        return 1
    }

    recipe_source="$source"
    [[ -f "$recipe_source" ]] || recipe_source="$(path_recipe)/${source}"
    if [[ -f "$recipe_source" && "$recipe_source" == *.y*ml ]]
    then
        inspect_recipe "$recipe_source"
        return
    fi

    dist_source="$source"
    [[ -d "$dist_source" ]] || dist_source="$(path_dist)/${source}"
    if [[ -d "$dist_source/.meta" && -f "$dist_source/.meta/recipe.yaml" ]]
    then
        inspect_dist "$dist_source"
        return
    fi

    if [[ -d "$source" ]]
    then
        inspect_ingredient "$source"
        return
    fi

    echo "inspect source not found: $source" >&2
    return 1
}

inspect_recipe() {
    local recipe_name="${1:-}"
    local output

    recipe_load "$recipe_name"
    recipe_name="$(recipe_name)"
    output="$(path_inspect)/recipe.${recipe_name}"
    FREEZE_WORKBENCH="$output"
    freeze_prepare
    freeze_collect_recipe
    inspect_generate_recipe_drafts
    rm -rf "${output}/ref" "${output}/lock" "${output}/recipe"
    FREEZE_WORKBENCH=""
    inspect_recipe_summary "${output}/summary.md"

    echo "[view] $output"
}

inspect_generate_recipe_drafts() {
    local workspace="$(freeze_workspace)"
    local recipe_dir="${workspace}/recipe"

    cp "$(recipe_file)" "${workspace}/recipe.draft.yaml"
    inspect_generate_ingredient_drafts "$recipe_dir" "$workspace" "$recipe_dir"
}

inspect_generate_ingredient_drafts() {
    local ingredient_dir="$1"
    local output_dir="$2"
    local header_root="$3"
    local extensions_draft="${output_dir}/extensions.draft"
    local settings_draft="${output_dir}/settings.draft"

    printf '# *** Inventory ***\n' > "$extensions_draft"
    find "$ingredient_dir" -type f \( -name '*.extensions' -o -name 'extensions' \) -print \
        | sort \
        | while IFS= read -r ingredient
        do
            echo "## ${ingredient#${header_root}/}"
            sed '/^$/d; s/^/+/' "$ingredient"
            echo
        done >> "$extensions_draft"

    printf '# *** Inventory ***\n' > "$settings_draft"
    find "$ingredient_dir" -type f \( -name '*.json' -o -name '*.jsonc' \) -print \
        | sort \
        | while IFS= read -r ingredient
        do
            echo "## ${ingredient#${header_root}/}"
            json_read "$ingredient" | json_gron | sed 's/^/+/'
            echo
        done >> "$settings_draft"
}

inspect_recipe_summary() {
    local output="$1"
    local profile profile_settings_default

    {
        echo "# Recipe: $(recipe_name)"
        echo
        echo "Source: \`$(recipe_file)\`"
        echo
        echo "## Ingredients"
        echo
        echo "- OS: $(recipe_os)"
        echo "- Platform: $(recipe_platform)"

        echo
        echo "### Runtime"
        recipe_runtimes | sed 's/^/- /'

        echo
        echo "### Profile"
        recipe_profiles | sed 's/^/- /'

        echo
        echo "## Resolved extensions"
        echo
        echo "### Runtime"
        case "$(recipe_default_profile_extensions)" in
        clean)
            echo '- (clean)'
            ;;
        runtime)
            resolve_runtime_extensions recipe \
                | extension_read recipe \
                | extension_merge \
                | sed 's/^/- /'
            ;;
        *)
            echo '- (invalid default-profile.extensions)'
            ;;
        esac

        while IFS= read -r profile
        do
            echo
            echo "### Profile: ${profile}"
            resolve_profile_extensions "$profile" recipe \
                | extension_read recipe \
                | extension_merge \
                | sed 's/^/- /'
        done < <(recipe_profiles)

        echo
        echo "## Resolved settings"
        echo
        echo '### Default'
        echo
        echo '```jsonc'
        resolve_settings | json_read | json_merge
        echo '```'

        while IFS= read -r profile
        do
            profile_settings_default="$(recipe_profile_default_flag "$profile" settings)" || return 1
            [[ "$profile_settings_default" == false ]] || continue
            echo
            echo "### Profile: ${profile}"
            echo
            echo '```jsonc'
            resolve_profile_settings "$profile" recipe | json_read | json_merge
            echo '```'
        done < <(recipe_profiles)
    } > "$output"
}

inspect_ingredient() {
    local ingredient_dir="${1:-$(path_ingredient)}"
    local output ingredient_root relative
    ingredient_root="$(path_ingredient)"
    ingredient_dir="$(cd "$ingredient_dir" && pwd)"

    [[ "$ingredient_dir" == "$ingredient_root" || "$ingredient_dir" == "${ingredient_root}/"* ]] || {
        echo "ingredient directory expected: $ingredient_dir" >&2
        return 1
    }

    relative="${ingredient_dir#${ingredient_root}/}"
    [[ "$relative" != "$ingredient_dir" ]] || relative=""
    relative="${relative//\//.}"
    output="$(path_inspect)/ingredient.${relative:-all}"

    rm -rf "$output"
    mkdir -p "$output"

    {
        echo "# Ingredient inventory: ${relative:-all}"
        echo
        find "$ingredient_dir" -type f ! -name '.*' -print \
            | sort \
            | while IFS= read -r ingredient
            do
                echo "- \`${ingredient#$ingredient_root/}\`"
            done
    } > "${output}/summary.md"

    inspect_generate_ingredient_drafts "$ingredient_dir" "$output" "$ingredient_root"

    echo "[view] $output"
}

inspect_dist() {
    local dist_name="${1:-}"
    local output lock_dir

    dist_load "$dist_name"
    dist_name="$(dist_name)"
    recipe_load "$(dist_recipe)" >/dev/null
    output="$(path_inspect)/dist.${dist_name}"

    FREEZE_WORKBENCH="$output"
    freeze_prepare
    freeze_collect_lock "$dist_name" false
    inspect_collect_empty_reference
    freeze_generate_extensions "$dist_name" "empty"
    freeze_generate_settings "$dist_name" "empty"
    freeze_generate_recipe
    rm -rf "${output}/ref" "${output}/lock" "${output}/recipe"
    FREEZE_WORKBENCH=""
    inspect_dist_summary "${output}/summary.md"

    echo "[view] $output"
}

inspect_collect_empty_reference() {
    local workspace="$(freeze_workspace)"

    printf '{}\n' > "${workspace}/ref/settings.jsonc"
    find "${workspace}/lock" -maxdepth 1 -type f \( -name '*.extensions' -o -name 'extensions' \) -print \
        | while IFS= read -r extension_file
        do
            : > "${workspace}/ref/$(basename "$extension_file")"
        done
}

inspect_dist_summary() {
    local output="$1"
    local lock_dir="${DIST_DIR}/.lock"

    {
        echo "# Dist: ${dist_name}"
        echo
        echo "Source: \`${DIST_DIR}\`"
        echo
        echo "## Recipe"
        echo
        echo '```yaml'
        cat "$(dist_recipe)"
        echo '```'

        echo
        echo "## Lock"
        echo
        if [[ ! -d "$lock_dir" ]]
        then
            echo "No Lock found."
        else
            echo "### Runtime extensions"
            echo
            echo '```text'
            cat "${lock_dir}/runtime.extensions.lock" 2>/dev/null || true
            echo '```'

            while IFS= read -r profile
            do
                echo
                echo "### Profile: ${profile}"
                echo
                echo '```text'
                cat "${lock_dir}/${profile}.extensions.lock" 2>/dev/null || true
                echo '```'
            done < <(recipe_profiles)

            echo
            echo "### Settings"
            echo
            echo "\`${lock_dir}/settings.jsonc\`"
        fi
    } > "$output"
}

inspect_sync() {
    local left="${1:-}"
    local right="${2:-}"
    local temp_root left_state right_state output

    [[ -n "$left" && -n "$right" ]] || {
        echo "usage: inspect sync <recipe.yaml | dist-dir> <recipe.yaml | dist-dir>" >&2
        return 1
    }

    inspect_resolve_sync_source "$left" || return 1
    local left_kind="$INSPECT_SOURCE_KIND"
    local left_path="$INSPECT_SOURCE_PATH"
    local left_label="$INSPECT_SOURCE_LABEL"

    inspect_resolve_sync_source "$right" || return 1
    local right_kind="$INSPECT_SOURCE_KIND"
    local right_path="$INSPECT_SOURCE_PATH"
    local right_label="$INSPECT_SOURCE_LABEL"

    temp_root="$(mktemp -d)"
    left_state="${temp_root}/left"
    right_state="${temp_root}/right"
    output="$(path_inspect)/sync.${left_label}.${right_label}"

    inspect_sync_collect_state "$left_kind" "$left_path" "$left_state" || {
        rm -rf "$temp_root"
        return 1
    }
    inspect_sync_collect_state "$right_kind" "$right_path" "$right_state" || {
        rm -rf "$temp_root"
        return 1
    }

    rm -rf "$output"
    mkdir -p "$output"
    inspect_sync_extensions "${left_state}/state" "${right_state}/state" "$left_label" "$right_label" \
        > "${output}/extensions.draft"
    inspect_sync_settings "${left_state}/state" "${right_state}/state" "$left_label" "$right_label" \
        > "${output}/settings.draft"
    inspect_sync_summary "$output" "$left_kind" "$left_path" "$left_label" \
        "$right_kind" "$right_path" "$right_label"

    rm -rf "$temp_root"
    echo "[sync] $output"
}

inspect_resolve_sync_source() {
    local source="${1:-}"
    local recipe_source dist_source label

    recipe_source="$source"
    [[ -f "$recipe_source" ]] || recipe_source="$(path_recipe)/${source}"
    if [[ -f "$recipe_source" && "$recipe_source" == *.y*ml ]]
    then
        INSPECT_SOURCE_KIND="recipe"
        INSPECT_SOURCE_PATH="$(cd "$(dirname "$recipe_source")" && pwd)/$(basename "$recipe_source")"
        label="$(basename "$recipe_source")"
        INSPECT_SOURCE_LABEL="${label%.*}"
        return
    fi

    dist_source="$source"
    [[ -d "$dist_source" ]] || dist_source="$(path_dist)/${source}"
    if [[ -d "$dist_source/.meta" && -f "$dist_source/.meta/recipe.yaml" ]]
    then
        INSPECT_SOURCE_KIND="dist"
        INSPECT_SOURCE_PATH="$(cd "$dist_source" && pwd)"
        INSPECT_SOURCE_LABEL="$(basename "$dist_source")"
        return
    fi

    echo "sync source not found (Recipe or Dist expected): $source" >&2
    return 1
}

inspect_sync_collect_state() {
    local kind="$1"
    local source="$2"
    local workspace="$3"
    local previous_workspace="$FREEZE_WORKBENCH"

    FREEZE_WORKBENCH="$workspace"
    freeze_prepare
    case "$kind" in
        recipe)
            recipe_file_set "$source" >/dev/null
            freeze_collect_reference >/dev/null
            mv "${workspace}/ref" "${workspace}/state"
            ;;
        dist)
            dist_load "$source" >/dev/null
            recipe_file_set "$(dist_recipe)" >/dev/null
            freeze_collect_lock "$(dist_name)" false
            mv "${workspace}/lock" "${workspace}/state"
            ;;
    esac
    FREEZE_WORKBENCH="$previous_workspace"
}

inspect_sync_extensions() {
    local left_state="$1"
    local right_state="$2"
    local left_label="$3"
    local right_label="$4"
    local extensions_file profile left_file right_file
    local difference=false

    printf '# *** Difference ***\n'
    while IFS= read -r extensions_file
    do
        left_file="${left_state}/${extensions_file}"
        right_file="${right_state}/${extensions_file}"
        [[ -f "$left_file" ]] || : > "$left_file"
        [[ -f "$right_file" ]] || : > "$right_file"

        if diff -q "$left_file" "$right_file" >/dev/null
        then
            continue
        fi

        difference=true
        if [[ "$extensions_file" == "runtime.extensions" ]]
        then
            echo "## runtime.draft.extensions"
        else
            profile="${extensions_file%.extensions}"
            echo "## profile/${profile}.extensions"
        fi
        echo
        echo '```diff'
        diff -u0bBN --suppress-common-lines \
            --label "$left_label" --label "$right_label" \
            "$left_file" "$right_file" || true
        echo '```'
        echo
    done < <(inspect_sync_extension_files "$left_state" "$right_state")

    if [[ "$difference" == false ]]
    then
        echo '[No differences] extensions'
    fi
}

inspect_sync_extension_files() {
    local left_state="$1"
    local right_state="$2"

    if [[ -f "${left_state}/runtime.extensions" || -f "${right_state}/runtime.extensions" ]]
    then
        echo runtime.extensions
    fi
    find "$left_state" "$right_state" -maxdepth 1 -type f -name '*.extensions' \
        ! -name 'runtime.extensions' -print \
        | while IFS= read -r path
          do
              basename "$path"
          done \
        | sort -u
}

inspect_sync_settings() {
    local left_state="$1"
    local right_state="$2"
    local left_label="$3"
    local right_label="$4"
    local settings_file left_file right_file header
    local difference=false

    printf '# *** Difference ***\n'
    while IFS= read -r settings_file
    do
        left_file="${left_state}/${settings_file}"
        right_file="${right_state}/${settings_file}"
        [[ -f "$left_file" ]] || : > "$left_file"
        [[ -f "$right_file" ]] || : > "$right_file"
        header="$(inspect_sync_settings_draft_header "$settings_file")" || continue

        if inspect_sync_settings_difference "$left_file" "$right_file" "$header" \
            "$left_label" "$right_label"
        then
            difference=true
        fi
    done < <(inspect_sync_settings_files "$left_state" "$right_state")

    if [[ "$difference" == false ]]
    then
        echo '[No differences] settings'
    fi
}

inspect_sync_settings_files() {
    local left_state="$1"
    local right_state="$2"

    if [[ -f "${left_state}/settings.jsonc" || -f "${right_state}/settings.jsonc" ]]
    then
        echo settings.jsonc
    fi
    find "$left_state" "$right_state" -maxdepth 1 -type f -name '*.settings.jsonc' -print \
        | while IFS= read -r path
          do
              basename "$path"
          done \
        | sort -u
}

inspect_sync_settings_draft_header() {
    local settings_file="$1"

    if [[ "$settings_file" == settings.jsonc ]]
    then
        echo 'runtime.draft.settings.json'
        return
    fi
    [[ "$settings_file" == *.settings.jsonc ]] || return 1
    echo "profile/${settings_file%.settings.jsonc}.settings.json"
}

inspect_sync_settings_difference() {
    local left_file="$1"
    local right_file="$2"
    local header="$3"
    local left_label="$4"
    local right_label="$5"
    local left_gron right_gron left_group right_group

    left_gron="$(mktemp)"
    right_gron="$(mktemp)"
    left_group="$(mktemp)"
    right_group="$(mktemp)"
    json_read "$left_file" | json_gron > "$left_gron"
    json_read "$right_file" | json_gron > "$right_gron"
    json_gron_group "$left_gron" > "$left_group"
    json_gron_group "$right_gron" > "$right_group"

    if diff -q "$left_group" "$right_group" >/dev/null
    then
        rm -f "$left_gron" "$right_gron" "$left_group" "$right_group"
        return 1
    fi

    {
        echo "## ${header}"
        echo
        echo '```diff'
        diff -u0bBN --suppress-common-lines \
            --label "$left_label" --label "$right_label" \
            "$left_group" "$right_group" || true
        echo '```'
    } | inspect_sync_format_settings_diff

    rm -f "$left_gron" "$right_gron" "$left_group" "$right_group"
}

inspect_sync_format_settings_diff() {
    local line content group current_group=""

    while IFS= read -r line
    do
        case "$line" in
            "@@ "*|"\\ No newline at end of file"|+|-)
                continue
                ;;
            "-// "*|"+// "*|"// "*)
                # Group comments are regenerated from changed settings below.
                continue
                ;;
            +json*|-json*)
                content="${line:1}"
                group="$(json_gron_group_name "$content")"
                if [[ "$group" != "$current_group" ]]
                then
                    [[ -n "$current_group" ]] && echo
                    echo "// ${group}"
                    current_group="$group"
                fi
                echo "$line"
                ;;
            '')
                ;;
            *)
                echo "$line"
                ;;
        esac
    done
}

inspect_sync_summary() {
    local output="$1"
    local left_kind="$2"
    local left_path="$3"
    local left_label="$4"
    local right_kind="$5"
    local right_path="$6"
    local right_label="$7"
    local left_recipe right_recipe

    left_recipe="$(inspect_sync_recipe_file "$left_kind" "$left_path")"
    right_recipe="$(inspect_sync_recipe_file "$right_kind" "$right_path")"

    {
        echo "# Sync: ${left_label} -> ${right_label}"
        echo
        echo "- Left (${left_kind}): \`${left_path}\`"
        echo "- Right (${right_kind}): \`${right_path}\`"
        echo
        echo 'Both sources are compared as resolved, completed states.'
        echo 'The Draft applies the left-to-right difference to Cookbook targets.'

        echo
        echo '## Left Recipe'
        echo
        echo '```yaml'
        cat "$left_recipe"
        echo '```'

        echo
        echo '## Right Recipe'
        echo
        echo '```yaml'
        cat "$right_recipe"
        echo '```'
    } > "${output}/summary.md"
}

inspect_sync_recipe_file() {
    local kind="$1"
    local source="$2"

    case "$kind" in
        recipe)
            echo "$source"
            ;;
        dist)
            echo "${source}/.lock/recipe.yaml"
            ;;
    esac
}
