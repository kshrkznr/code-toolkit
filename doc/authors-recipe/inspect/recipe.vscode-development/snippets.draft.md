# Snippets View

## Inventory

### runtime.draft.snippets.shellscript.json

```diff
+ {
+   "Add Function": {
+     "body": [
+       "${1:function_name}() {",
+       "    local arg=\\${1:-}",
+       "    $0",
+       "}"
+     ],
+     "description": "Add a Bash function with an optional first argument",
+     "prefix": "func"
+   },
+   "New Main": {
+     "body": [
+       "#!/usr/bin/env bash",
+       "set -euo pipefail",
+       "",
+       "main() {",
+       "    $0",
+       "}",
+       "",
+       "main \"\\$@\""
+     ],
+     "description": "Create a strict Bash script with a main function",
+     "prefix": [
+       "new",
+       "main"
+     ]
+   },
+   "New Script": {
+     "body": [
+       "#!/usr/bin/env bash",
+       "$0"
+     ],
+     "description": "Create a minimal Bash script",
+     "prefix": [
+       "new",
+       "script"
+     ]
+   },
+   "Stdin or Args": {
+     "body": [
+       "stdin_or_args() {",
+       "    if ((\\$#)); then",
+       "        printf '%s\\n' \"\\$@\"",
+       "    else",
+       "        cat",
+       "    fi",
+       "}"
+     ],
+     "description": "Read records from arguments or standard input",
+     "prefix": "stdin"
+   },
+   "filter_non_empty": {
+     "body": [
+       "filter_non_empty() {",
+       "    while IFS= read -r line || [[ -n \\$line ]]; do",
+       "        [[ -n \\$line ]] \u0026\u0026 printf '%s\\n' \"\\$line\"",
+       "    done",
+       "}"
+     ],
+     "description": "Remove empty records from standard input",
+     "prefix": "filter"
+   },
+   "map_lines": {
+     "body": [
+       "map_lines() {",
+       "    local func=\"\\$1\"",
+       "    shift",
+       "",
+       "    stdin_or_args \"\\$@\" |",
+       "    while IFS= read -r line || [[ -n \\$line ]]; do",
+       "        \"\\$func\" \"\\$line\"",
+       "    done",
+       "}"
+     ],
+     "description": "Apply a function to each input record",
+     "prefix": "lines"
+   }
+ }
```
