#!/bin/bash
# Script to analyze Gno packages and generate a markdown report
# Usage: ./analyze-packages.sh [directory] [filter]
# Example: ./analyze-packages.sh examples/gno.land
# Example: ./analyze-packages.sh examples/gno.land "p/demo"
# Example: ./analyze-packages.sh examples/gno.land "moul"
# Make sure to run this from the misc/ folder
# since EXAMPLES_DIR is relative

DIR="${1:-examples/gno.land}"
FILTER="${2:-}"

# Temporary files for data collection
TEMP_DATA=$(mktemp)
TEMP_HEADERS=$(mktemp)
trap "rm -f $TEMP_DATA $TEMP_HEADERS" EXIT

# Function to count non-comment lines in non-test .gno files
count_lines() {
    local dir="$1"
    
    # Find all .gno files that are not test files and count non-comment lines
    find "$dir" -name "*.gno" -type f ! -name "*_test.gno" 2>/dev/null | while read -r file; do
        # Count non-empty, non-comment lines (excluding // comments)
        grep -v '^[[:space:]]*\/\/' "$file" 2>/dev/null | grep -v '^[[:space:]]*$'
    done | wc -l
}

# Function to find packages that import this one
count_dependents() {
    local pkg="$1"
    local base_dir="$2"
    
    # Count how many times this package is imported
    local pkg_short="${pkg#gno.land/}"
    local count=$(grep -r "import.*\"$pkg\"" "$base_dir" --include="*.gno" 2>/dev/null | \
        grep -v "^$pkg_short/" | \
        cut -d: -f1 | \
        sort -u | \
        wc -l)
    
    echo "$count"
}

# Function to check if README exists
has_readme() {
    local dir="$1"
    if [ -f "$dir/README.md" ] || [ -f "$dir/README" ] || [ -f "$dir/readme.md" ]; then
        echo "yes"
    else
        echo "no"
    fi
}

# Function to extract flags from gnomod.toml
get_flags() {
    local gnomod="$1"
    local flags=""
    
    if [ -f "$gnomod" ]; then
        # Check for draft flag
        if grep -q "^draft = true" "$gnomod" 2>/dev/null; then
            flags="${flags}draft,"
        fi
        
        # Check for private flag
        if grep -q "^private = true" "$gnomod" 2>/dev/null; then
            flags="${flags}private,"
        fi
        
        # Check for ignore flag
        if grep -q "^ignore = true" "$gnomod" 2>/dev/null; then
            flags="${flags}ignore,"
        fi
    fi
    
    # Remove trailing comma
    echo "${flags%,}"
}

# Function to get first git committer and date
get_first_git_info() {
    local dir="$1"
    local info_type="$2"  # "author" or "date"
    
    # Get the first commit info for files in this directory
    local format="%an"
    if [ "$info_type" = "date" ]; then
        format="%ad"
    elif [ "$info_type" = "author" ]; then
        # Use email format to extract GitHub handle
        format="%ae"
    fi
    
    # Use git log with path to be more efficient
    local info=$(git log --follow --format="$format" --date=short --reverse -- "$dir/*.gno" 2>/dev/null | head -1)
    
    if [ -z "$info" ]; then
        # Fallback: try to find any .gno file
        local gno_file=$(find "$dir" -name "*.gno" -type f | head -1)
        if [ -n "$gno_file" ]; then
            info=$(git log --follow --format="$format" --date=short --reverse -- "$gno_file" 2>/dev/null | head -1)
        fi
    fi
    
    # Process author info to get GitHub handle
    if [ "$info_type" = "author" ]; then
        # Try to extract GitHub handle from email or use the email prefix
        if [[ "$info" =~ ^([^@]+)@users.noreply.github.com$ ]]; then
            # Extract GitHub handle from noreply email
            info="${BASH_REMATCH[1]}"
            # Remove any numeric prefix (like "123456+")
            info=$(echo "$info" | sed 's/^[0-9]*+//')
        elif [[ "$info" =~ ^([^@]+)@ ]]; then
            # Use email prefix as fallback
            info="${BASH_REMATCH[1]}"
        fi
        info="${info:-unknown}"
    else
        info="${info:-unknown}"
    fi
    
    echo "$info"
}

# Collect all data first
echo "# Gno Packages Analysis"
echo ""
echo "Generated on: $(date)"
echo ""

# Find all directories with gnomod.toml and collect data
find "$DIR" -name "gnomod.toml" -type f | sort | while read -r gnomod_path; do
    pkg_dir=$(dirname "$gnomod_path")
    
    # Extract package path from gnomod.toml
    if [ -f "$gnomod_path" ]; then
        module_line=$(grep "^module = " "$gnomod_path" | head -1)
        if [ -n "$module_line" ]; then
            # Extract module name between quotes
            pkg_path=$(echo "$module_line" | sed 's/module = "\(.*\)"/\1/')
            
            # Remove gno.land/ prefix for display
            display_path="${pkg_path#gno.land/}"
            
            # Apply filter if provided
            if [ -n "$FILTER" ]; then
                # Skip if path doesn't match filter
                if ! echo "$display_path" | grep -q "$FILTER"; then
                    # Also check if author matches filter
                    temp_author=$(get_first_git_info "$pkg_dir" "author")
                    if ! echo "$temp_author" | grep -qi "$FILTER"; then
                        continue
                    fi
                fi
            fi
            
            # Collect all data
            line_count=$(count_lines "$pkg_dir")
            dependent_count=$(count_dependents "$pkg_path" "$DIR")
            has_readme_val=$(has_readme "$pkg_dir")
            flags=$(get_flags "$gnomod_path")
            flags="${flags:-none}"
            first_author=$(get_first_git_info "$pkg_dir" "author")
            first_date=$(get_first_git_info "$pkg_dir" "date")
            
            # Store data in temporary file
            echo "$display_path|$line_count|$dependent_count|$has_readme_val|$flags|$first_author|$first_date" >> "$TEMP_DATA"
        fi
    fi
done

# Determine which columns to show
show_dependents=false
show_readme=false
show_flags=false

if [ -s "$TEMP_DATA" ]; then
    # Check if any dependents > 0
    while IFS='|' read -r pkg lines deps readme flags author date; do
        if [ "$deps" -gt 0 ] 2>/dev/null; then
            show_dependents=true
            break
        fi
    done < "$TEMP_DATA"
    
    # Check if any README = yes
    while IFS='|' read -r pkg lines deps readme flags author date; do
        if [ "$readme" = "yes" ]; then
            show_readme=true
            break
        fi
    done < "$TEMP_DATA"
    
    # Check if any flags != none
    while IFS='|' read -r pkg lines deps readme flags author date; do
        if [ "$flags" != "none" ] && [ -n "$flags" ]; then
            show_flags=true
            break
        fi
    done < "$TEMP_DATA"
fi

# Build header and separator
header="| Package | Lines"
separator="|---------|-------"

if [ "$show_dependents" = true ]; then
    header="$header | Dependents"
    separator="$separator |------------"
fi

if [ "$show_readme" = true ]; then
    header="$header | README"
    separator="$separator |--------"
fi

if [ "$show_flags" = true ]; then
    header="$header | Flags"
    separator="$separator |-------"
fi

header="$header | First Author | First Date |"
separator="$separator |--------------|------------|"

echo "$header"
echo "$separator"

# Output data rows
if [ -s "$TEMP_DATA" ]; then
    while IFS='|' read -r pkg lines deps readme flags author date; do
        row="| $pkg | $lines"
        
        if [ "$show_dependents" = true ]; then
            # Hide default value (0)
            if [ "$deps" = "0" ]; then
                row="$row | "
            else
                row="$row | $deps"
            fi
        fi
        
        if [ "$show_readme" = true ]; then
            # Hide default value (no)
            if [ "$readme" = "no" ]; then
                row="$row | "
            else
                row="$row | $readme"
            fi
        fi
        
        if [ "$show_flags" = true ]; then
            # Hide default value (none)
            if [ "$flags" = "none" ]; then
                row="$row | "
            else
                row="$row | $flags"
            fi
        fi
        
        row="$row | $author | $date |"
        echo "$row"
    done < "$TEMP_DATA"
fi

# Summary statistics
echo ""
echo "## Summary Statistics"
echo ""

# Count total packages
total_packages=$(find "$DIR" -name "gnomod.toml" -type f | wc -l)
echo "- Total packages: $total_packages"

# Count packages by type
p_packages=$(find "$DIR/p" -name "gnomod.toml" -type f 2>/dev/null | wc -l)
r_packages=$(find "$DIR/r" -name "gnomod.toml" -type f 2>/dev/null | wc -l)
echo "- Pure packages (p/): $p_packages"
echo "- Realm packages (r/): $r_packages"

# Count total lines
total_lines=$(count_lines "$DIR")
echo "- Total lines of code: $total_lines"

# Count packages with flags
draft_count=$(grep -l "^draft = true" $(find "$DIR" -name "gnomod.toml" -type f) 2>/dev/null | wc -l)
private_count=$(grep -l "^private = true" $(find "$DIR" -name "gnomod.toml" -type f) 2>/dev/null | wc -l)
ignore_count=$(grep -l "^ignore = true" $(find "$DIR" -name "gnomod.toml" -type f) 2>/dev/null | wc -l)

echo "- Packages marked as draft: $draft_count"
echo "- Packages marked as private: $private_count"
echo "- Packages marked as ignore: $ignore_count"