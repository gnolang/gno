#!/bin/bash
# Script to analyze package-level changes between current branch and master
# Usage: ./analyze-package-changes.sh [base-branch]
# Example: ./analyze-package-changes.sh master

BASE_BRANCH="${1:-master}"
EXAMPLES_DIR="examples"

# Temporary files
TEMP_BASE=$(mktemp)
TEMP_CURRENT=$(mktemp)
TEMP_RENAMES=$(mktemp)
trap "rm -f $TEMP_BASE $TEMP_CURRENT $TEMP_RENAMES" EXIT

echo "# Package Changes Analysis"
echo ""
echo "Comparing current branch with: $BASE_BRANCH"
echo "Generated on: $(date)"
echo ""

# Get all packages (directories with .gno files) in base branch
git ls-tree -r --name-only "$BASE_BRANCH" | grep "^$EXAMPLES_DIR/.*\.gno$" | while read -r file; do
    dirname "$file"
done | sort -u > "$TEMP_BASE"

# Get all packages in current branch
find "$EXAMPLES_DIR" -name "*.gno" -type f 2>/dev/null | while read -r file; do
    dirname "$file"
done | sort -u > "$TEMP_CURRENT"

# Find renamed packages using git log
echo "## Renamed Packages"
echo ""

# Get rename information from git
git diff --name-status --diff-filter=R "$BASE_BRANCH"..HEAD -- "$EXAMPLES_DIR" | grep "\.gno$" | while read -r status old_file new_file; do
    old_pkg=$(dirname "$old_file")
    new_pkg=$(dirname "$new_file")
    if [ "$old_pkg" != "$new_pkg" ]; then
        echo "$old_pkg -> $new_pkg" >> "$TEMP_RENAMES"
    fi
done

# Sort and deduplicate renames
if [ -s "$TEMP_RENAMES" ]; then
    sort -u "$TEMP_RENAMES" | while read -r rename; do
        old_pkg=$(echo "$rename" | cut -d' ' -f1)
        new_pkg=$(echo "$rename" | cut -d' ' -f3)
        
        # Check if content was modified beyond just renaming
        modified=""
        
        # Check if package has any modifications besides renames
        # Look for any Modified (M) or Added (A) files in the new package location
        if git diff --name-status "$BASE_BRANCH"..HEAD -- "$new_pkg" | grep -E "^[MA].*\.gno$" -q; then
            modified=" (modified)"
        else
            # Also check if renamed files have content changes
            git diff --name-status --diff-filter=R "$BASE_BRANCH"..HEAD | grep "\.gno$" | while read -r status old_file new_file; do
                if [[ "$old_file" == "$old_pkg"/* ]] && [[ "$new_file" == "$new_pkg"/* ]]; then
                    # Check if the renamed file has content changes
                    if ! git diff "$BASE_BRANCH:$old_file" "HEAD:$new_file" --quiet 2>/dev/null; then
                        return 0  # Signal that we found changes
                    fi
                fi
            done
            if [ $? -eq 0 ]; then
                modified=" (modified)"
            fi
        fi
        
        echo "- \`$old_pkg\` → \`$new_pkg\`$modified"
    done
else
    echo "*No packages renamed*"
fi

echo ""
echo "## Deleted Packages"
echo ""

# Find deleted packages
deleted_count=0
while read -r pkg; do
    # Check if package exists in current branch
    if ! grep -q "^$pkg$" "$TEMP_CURRENT"; then
        # Check if it was renamed (already handled above)
        if ! grep -q "^$pkg " "$TEMP_RENAMES"; then
            echo "- \`$pkg\`"
            ((deleted_count++))
        fi
    fi
done < "$TEMP_BASE"

if [ "$deleted_count" -eq 0 ]; then
    echo "*No packages deleted*"
fi

echo ""
echo "## New Packages"
echo ""

# Find new packages
new_count=0
while read -r pkg; do
    # Check if package exists in base branch
    if ! grep -q "^$pkg$" "$TEMP_BASE"; then
        # Check if it was renamed (already handled above)
        if ! grep -q " $pkg$" "$TEMP_RENAMES"; then
            echo "- \`$pkg\`"
            ((new_count++))
        fi
    fi
done < "$TEMP_CURRENT"

if [ "$new_count" -eq 0 ]; then
    echo "*No new packages added*"
fi

echo ""
echo "## Modified Packages"
echo ""
echo "(Packages with content changes, excluding renames)"
echo ""

# Find modified packages (content changed but not renamed or deleted)
modified_count=0
git diff --name-status "$BASE_BRANCH"..HEAD -- "$EXAMPLES_DIR" | grep "\.gno$" | while read -r status file rest; do
    if [ "$status" = "M" ]; then
        pkg=$(dirname "$file")
        # Only show each package once
        echo "$pkg"
    fi
done | sort -u | while read -r pkg; do
    # Skip if already shown as renamed
    if ! grep -q "$pkg$" "$TEMP_RENAMES" && ! grep -q "^[^ ]* $pkg$" "$TEMP_RENAMES"; then
        echo "- \`$pkg\`"
        ((modified_count++))
    fi
done

if [ "$modified_count" -eq 0 ]; then
    # Check if we actually had any output
    if ! git diff --name-status "$BASE_BRANCH"..HEAD -- "$EXAMPLES_DIR" | grep -q "^M.*\.gno$"; then
        echo "*No packages modified*"
    fi
fi

echo ""
echo "## Summary Statistics"
echo ""

# Count statistics
renamed_count=$([ -s "$TEMP_RENAMES" ] && sort -u "$TEMP_RENAMES" | wc -l || echo 0)
total_base=$(wc -l < "$TEMP_BASE")
total_current=$(wc -l < "$TEMP_CURRENT")

echo "- Total packages in $BASE_BRANCH: $total_base"
echo "- Total packages in current branch: $total_current"
echo "- Packages renamed: $renamed_count"
echo "- Packages deleted: $deleted_count"
echo "- Packages added: $new_count"
echo "- Net change: $((total_current - total_base)) packages"