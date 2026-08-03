#!/usr/bin/env bash

path_platform_home(){
    local platform="${1:-code}"
    case "$platform" in
        code) echo "$HOME/AppData/Roaming/Code" ;;
        kiro) echo "$HOME/AppData/Roaming/Kiro" ;;
        *) return 1 ;;
    esac
}

path_platform_extHome(){
    local platform="${1:-code}"
    case "$platform" in
        code) echo "$HOME/.vscode" ;;
        kiro) echo "$HOME/.kiro" ;;
        *) return 1 ;;
    esac
}
