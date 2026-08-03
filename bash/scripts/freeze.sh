#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/lifecycle/freeze.sh"
source "${SCRIPT_DIR}/lib/workbench.sh"

main() {
    local mode="${1:-}"
    local phase="${2:-}"
    local dist_name="${3:-}"

    case "$mode" in
        modelist)
            modelist
            ;;
		freeze)
			freeze "$phase" "$dist_name"
			;;
		workbench)
			shift
			open_workbench "$@"
			;;
        *)
            usage
            exit 1
            ;;
    esac
}

freeze (){
    local phase="${1:-}"
    local dist_name="${2:-}"
    if [[ -z "$phase" ]]
    then
        phase="$(freeze_phaselist | fzf)"
    fi
    case "$phase" in
        draft)
            freeze_draft "$dist_name"
            ;;
        edit)
            freeze_edit "${2:-}"
            ;;
        commit)
            freeze_drafts_commit
            ;;
        *)
            echo "phase : $(freeze_phaselist)"
            exit 1
            ;;
    esac
}

modelist() {
	printf '%s\n' freeze workbench
}

open_workbench() {
	local kind=""
	local viewpoint=""
	local editor=""
	while (($#))
	do
		case "$1" in
			--editor)
				shift
				[[ $# -gt 0 && -n "$1" ]] || {
					echo "--editor requires a command" >&2
					return 1
				}
				editor="$1"
				;;
			--*)
				echo "unknown workbench option: $1" >&2
				return 1
				;;
			*)
				if [[ -z "$kind" ]]
				then
					kind="$1"
				elif [[ -z "$viewpoint" ]]
				then
					viewpoint="$1"
				else
					echo "too many workbench arguments" >&2
					return 1
				fi
				;;
		esac
		shift
	done

	if [[ -z "$kind" ]]
	then
		kind="$(workbench_kinds | fzf)"
		[[ -n "$kind" ]] || return 0
	fi

	local target
	case "$kind" in
		draft)
			[[ -z "$viewpoint" ]] || {
				echo "draft Workbench does not accept a viewpoint" >&2
				return 1
			}
			target="$(path_draft)"
			;;
		inspect)
			if [[ -z "$viewpoint" ]]
			then
				viewpoint="$(inspect_workbenches | fzf)"
				[[ -n "$viewpoint" ]] || return 0
			fi
			inspect_workbenches | grep -Fqx -- "$viewpoint" || {
				echo "Inspect Workbench not found: $viewpoint" >&2
				return 1
			}
			target="$(path_inspect)/${viewpoint}"
			;;
		*)
			echo "workbench must be draft or inspect: $kind" >&2
			return 1
			;;
	esac

	[[ -d "$target" ]] || {
		echo "Workbench not found: $target" >&2
		return 1
	}
	workbench_open_path "$target" "$editor"
}

workbench_kinds() {
	[[ -d "$(path_draft)" ]] && echo draft
	[[ -n "$(inspect_workbenches)" ]] && echo inspect
}

inspect_workbenches() {
	local root
	root="$(path_inspect)"
	[[ -d "$root" ]] || return 0
	find "$root" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort
}

freeze_phaselist() {
    echo draft
    echo edit
    echo commit
}

freeze_draft() {
    freeze_prepare
    local dist_name="${1:-}"
    dist_load "${dist_name}"
    dist_name=$(dist_name)
    freeze_collect_lock "$dist_name"

    recipe_load "$(dist_recipe)"
    resolve_load_recipe "$(recipe_name)" "$(recipe_os)" > /dev/null
    echo "recipe resolved : $(recipe_file)"
    local recipe_name="$(recipe_name)"
    freeze_collect_recipe
    freeze_collect_reference   "$dist_name"
    freeze_generate_extensions "$dist_name" "$recipe_name"
    freeze_generate_settings   "$dist_name" "$recipe_name"
    freeze_generate_recipe     "$dist_name"
    echo
    echo "[done] freeze draft $(dist_name)"
}

freeze_edit() {
	local editor
	editor="$(workbench_editor "${1:-}")"
	while read -r draft
	do
		[[ -f "$draft" ]] || continue
		workbench_open_path "$draft" "$editor" background
    done < <(find_draft)
    echo
    echo "[done] freeze edit"
}

freeze_drafts_commit() {
    while read -r recipe
    do
        [[ -f "$recipe" ]] || continue
        freeze_commit_recipe "$recipe"
    done < <({
        find_file_thin "$(path_draft)" "*recipe.yaml"
        find_file_thin "$(path_draft)" "*recipe.draft.yaml"
        find_file_thin "$(path_draft)" "*recipe.yml"
        find_file_thin "$(path_draft)" "*recipe.draft.yml"
    } | sort -u)

    while read -r draft
    do
        [[ -f "$draft" ]] || continue
        freeze_commit "$draft"
    done < <(find_draft)
    echo
    echo "[done] freeze commit"
}

main "$@"
