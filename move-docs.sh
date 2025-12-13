#!/bin/bash
# Move documentation files to docs folder

cd "$(dirname "$0")"

# Files to move to docs/
FILES_TO_DOCS=(
    "BROWSER_EXTENSIONS.md"
    "BROWSER_EXTENSION_IMPLEMENTATION.md"
    "NETWORK_INFO.md"
    "SECURITY_IMPROVEMENTS.md"
    "CONTRIBUTING.md"
)

for file in "${FILES_TO_DOCS[@]}"; do
    if [ -f "$file" ]; then
        echo "Moving $file to docs/"
        mv "$file" "docs/"
    fi
done

echo "Done! Documentation files moved to docs/"
ls -la docs/

