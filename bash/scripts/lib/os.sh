#!/usr/bin/env bash
case "$(uname -s)" in
    Darwin)
        source "${SCRIPT_DIR}/lib/macos/path.sh"
        source "${SCRIPT_DIR}/lib/macos/cmd.sh"
        ;;
    MINGW*|MSYS*)
        export PATH="/usr/bin:$PATH"
        source "${SCRIPT_DIR}/lib/windows/path.sh"
        source "${SCRIPT_DIR}/lib/windows/cmd.sh"
        ;;
    *)
        echo "Unsupported platform"
        exit 1
        ;;
esac
