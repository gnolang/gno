#!/bin/bash
# Update gas-wanted values in txtar integration tests to 110% of actual usage.
#
# Usage:
#   cd gno
#   ./gno.land/pkg/integration/update_gas_wanted.sh
#
# This runs the integration tests with verbose output, captures GAS USED values,
# computes new gas-wanted (110% rounded to 2 significant digits with _ separators),
# rewrites the txtar files, then runs UPDATE_SCRIPTS=true to fix stdout assertions.

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

TESTDATA="$SCRIPT_DIR/testdata"
OUTPUT=$(mktemp)

echo "Step 1: Running integration tests to capture gas values..."
INMEMORY_TS=true go test -v -p 1 -timeout=30m ./gno.land/pkg/integration/ -run TestTestdata -count=1 > "$OUTPUT" 2>&1 || true

echo "Step 2: Parsing gas values and rewriting txtar files..."
python3 -c "
import re, os, sys, glob

def round_2sig(val):
    \"\"\"Round to nearest 2 significant digits: 15658754 -> 16000000\"\"\"
    if val <= 0:
        return 100000
    s = str(val)
    n = len(s)
    if n <= 2:
        return val
    first2 = int(s[:2])
    remainder = int(s[2:]) if n > 2 else 0
    half = 5 * (10 ** (n - 3)) if n > 2 else 0
    if remainder >= half:
        first2 += 1
    if first2 >= 100:
        first2 = 10
        n += 1
    return first2 * (10 ** (n - 2))

def format_us(val):
    s = str(val)
    parts = []
    while len(s) > 3:
        parts.append(s[-3:])
        s = s[:-3]
    parts.append(s)
    return '_'.join(reversed(parts))

# Parse test output
with open('$OUTPUT') as f:
    content = f.read()

# Extract per-test gas pairs. Since -p 1, tests run sequentially.
# But subtests may interleave, so use === RUN markers.
current_test = None
test_gas = {}
pending_wanted = None

for line in content.split('\n'):
    m = re.search(r'=== RUN\s+TestTestdata/(\S+)', line)
    if m:
        current_test = m.group(1)
        if current_test not in test_gas:
            test_gas[current_test] = []
        pending_wanted = None
        continue
    if not current_test:
        continue
    if '> stdout' in line or '> stderr' in line:
        continue
    m = re.search(r'GAS WANTED:\s+(\d+)', line)
    if m:
        pending_wanted = int(m.group(1))
    m = re.search(r'GAS USED:\s+(\d+)', line)
    if m and pending_wanted is not None:
        test_gas[current_test].append((pending_wanted, int(m.group(1))))
        pending_wanted = None

# Process each txtar file
stats = {'files': 0, 'commands': 0}
for fpath in sorted(glob.glob('$TESTDATA/*.txtar')):
    basename = os.path.basename(fpath).replace('.txtar', '')
    gas_pairs = test_gas.get(basename, [])
    if not gas_pairs:
        continue

    with open(fpath) as f:
        lines = f.readlines()

    changed = False
    gas_idx = 0

    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped.startswith('#') or stripped.startswith('stdout') or stripped.startswith('stderr'):
            continue

        m = re.search(r'gas-wanted\s+(\d[\d_]*)', line)
        if not m:
            continue

        old_gw_str = m.group(1)
        old_gw = int(old_gw_str.replace('_', ''))

        if gas_idx < len(gas_pairs):
            test_wanted, actual_used = gas_pairs[gas_idx]
            if test_wanted == old_gw:
                gas_idx += 1
                # Skip intentional OOG tests
                if actual_used > test_wanted:
                    continue
                new_gw = round_2sig(int(actual_used * 1.1))
                new_gw_str = format_us(new_gw)
                if old_gw_str.replace('_', '') != str(new_gw):
                    new_line = line.replace(f'gas-wanted {old_gw_str}', f'gas-wanted {new_gw_str}')
                    # Fix gas-fee
                    fee_m = re.search(r'gas-fee\s+(\d[\d_]*)ugnot', new_line)
                    if fee_m:
                        new_fee = max(int(new_gw * 0.1) + 1, 1000)
                        new_line = re.sub(r'gas-fee\s+\d[\d_]*ugnot', f'gas-fee {new_fee}ugnot', new_line)
                    lines[i] = new_line
                    changed = True
                    stats['commands'] += 1
                    # Fix GAS WANTED assertion
                    for j in range(i+1, min(i+15, len(lines))):
                        gw_assert = re.search(r\"GAS WANTED:\s+(\d+)\", lines[j])
                        if gw_assert:
                            old_assert = gw_assert.group(1)
                            if old_assert != str(new_gw):
                                lines[j] = lines[j].replace(f'GAS WANTED: {old_assert}', f'GAS WANTED: {new_gw}')
                            break
            else:
                # Mismatch — likely a ! command with no gas output, skip
                pass
        # else: no more gas data

    if changed:
        with open(fpath, 'w') as f:
            f.writelines(lines)
        stats['files'] += 1

print(f\"Updated {stats['files']} files, {stats['commands']} commands\")
"

echo "Step 3: Running UPDATE_SCRIPTS=true to fix remaining assertions..."
INMEMORY_TS=true UPDATE_SCRIPTS=true go test -v -p 1 -timeout=30m ./gno.land/pkg/integration/ -run TestTestdata -count=1 > /dev/null 2>&1 || true

echo "Step 4: Verifying all tests pass..."
INMEMORY_TS=true go test -p 1 -timeout=30m ./gno.land/pkg/integration/ -run TestTestdata -count=1

rm -f "$OUTPUT"
echo "Done!"
