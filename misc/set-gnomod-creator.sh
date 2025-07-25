#!/bin/sh
# Set creator address in gnomod.toml files
# Usage: ./set-gnomod-creator.sh <path> <address>
# Example: ./set-gnomod-creator.sh ../examples/gno.land/p/demo g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5

[ $# -ne 2 ] && echo "Usage: $0 <path> <address>" && exit 1

find "$1" -name "gnomod.toml" 2>/dev/null | while read -r file; do
    if ! grep -q "addpkg" "$file"; then
        printf "\n[addpkg]\ncreator = \"%s\"\n" "$2" >> "$file"
        echo "Updated: $file"
    fi
done