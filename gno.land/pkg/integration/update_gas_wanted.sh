#!/bin/bash
# Update gas-wanted values in txtar integration tests to round_2sig(actual * 1.1).
#
# Usage:
#   cd gno
#   ./gno.land/pkg/integration/update_gas_wanted.sh
#
# How it works:
#   1. Snapshot every testdata/*.txtar to a tmpdir (trap EXIT restores it).
#   2. Inflate every -gas-wanted N to N*10 in the live testdata files. Skip
#      lines that are ! gnokey (negative tests) or that are preceded by a
#      '# gas-rewrite: skip' marker comment. gas-fee is NOT touched (so we
#      preserve the gas-price ratio and the chain's insufficient-fee check
#      stays satisfied for the inflated run).
#   3. Run the integration tests with the inflated values. With 10x headroom
#      we capture true GAS USED for every command that previously OOG'd.
#   4. Restore the originals from the tmpdir snapshot (BEFORE parsing, so a
#      Python crash can't leave the working tree inflated).
#   5. Parse captured output: per-test (test_wanted, actual_used) pairs.
#      pending_wanted is reset on every '> gnokey' marker so that a tx that
#      prints GAS WANTED but no GAS USED (mid-tx error) doesn't bind to the
#      next command's GAS USED.
#   6. Walk each restored txtar. For each non-skipped, non-! gnokey line:
#      match the captured pair where test_wanted == old_gw * 10 (because we
#      inflated). Compute new_gw = round_2sig(actual * 1.1). Apply an
#      idempotence band: skip if old_gw is in [new_gw, new_gw*2]. Leave
#      gas-fee untouched. Update the matching 'GAS WANTED: N' assertion in
#      the command's region.
#   7. Run UPDATE_SCRIPTS=true to refresh cmp golden files.
#   8. Verify: re-run the suite and require it to pass.
#
# Properties:
#   - Idempotent: actual gas usage is determined by tx contents, so re-runs
#     produce the same new_gw and the band guard absorbs sub-1% noise.
#   - Robust to gas-cost code changes: inflated Step 1 won't OOG.
#   - Doesn't churn balance/'TOTAL TX COST' assertions (gas-fee untouched).

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

TESTDATA="$SCRIPT_DIR/testdata"
SNAPSHOT_DIR="$(mktemp -d -t update_gas_wanted_snapshot.XXXXXX)"
OUTPUT="$(mktemp -t update_gas_wanted_output.XXXXXX)"
INFLATION_FACTOR=10

# Always restore originals from snapshot, even on Ctrl-C / kill / error.
restore_and_cleanup() {
    if [ -d "$SNAPSHOT_DIR" ] && [ -n "$(ls -A "$SNAPSHOT_DIR" 2>/dev/null)" ]; then
        rsync -a --delete "$SNAPSHOT_DIR/" "$TESTDATA/"
    fi
    rm -rf "$SNAPSHOT_DIR"
    rm -f "$OUTPUT"
}
trap restore_and_cleanup EXIT

echo "Step 0: Snapshotting testdata to $SNAPSHOT_DIR..."
rsync -a "$TESTDATA/" "$SNAPSHOT_DIR/"

echo "Step 1: Inflating gas-wanted values by ${INFLATION_FACTOR}x..."
INFLATION_FACTOR=$INFLATION_FACTOR TESTDATA="$TESTDATA" python3 - <<'PYEOF'
import os, re, glob

INFLATION = int(os.environ['INFLATION_FACTOR'])
TESTDATA = os.environ['TESTDATA']

# Walk back past blank lines AND any '#'-prefixed comment lines to find a
# 'gas-rewrite: skip' marker. The marker is honored if it appears anywhere
# in the contiguous comment/blank block immediately above a command line.
def is_skipped(lines, i):
    k = i - 1
    while k >= 0:
        s = lines[k].strip()
        if s == '':
            k -= 1
            continue
        if s.startswith('#'):
            if re.match(r'#\s*gas-rewrite:\s*skip\b', s):
                return True
            k -= 1
            continue
        break
    return False

def format_us(val):
    s = str(val)
    parts = []
    while len(s) > 3:
        parts.append(s[-3:])
        s = s[:-3]
    parts.append(s)
    return '_'.join(reversed(parts))

n_files = 0
n_lines = 0
for fpath in sorted(glob.glob(os.path.join(TESTDATA, '*.txtar'))):
    with open(fpath, encoding='utf-8') as f:
        lines = f.readlines()
    changed = False
    for i, line in enumerate(lines):
        # Don't inflate negative-test commands; their tightness is the test.
        if re.match(r'\s*!\s+gnokey\b', line):
            continue
        # Don't inflate explicitly-skipped commands.
        if is_skipped(lines, i):
            continue
        m = re.search(r'(-gas-wanted\s+)(\d[\d_]*)', line)
        if not m:
            continue
        old = int(m.group(2).replace('_', ''))
        new = old * INFLATION
        new_str = format_us(new) if '_' in m.group(2) else str(new)
        lines[i] = line[:m.start(2)] + new_str + line[m.end(2):]
        changed = True
        n_lines += 1
    if changed:
        with open(fpath, 'w', encoding='utf-8') as f:
            f.writelines(lines)
        n_files += 1

print(f"  inflated {n_lines} lines across {n_files} files")
PYEOF

echo "Step 2: Running integration tests with inflated values..."
INMEMORY_TS=true go test -v -p 1 -timeout=30m ./gno.land/pkg/integration/ -run TestTestdata -count=1 > "$OUTPUT" 2>&1 || true

echo "Step 3: Restoring testdata from snapshot..."
rsync -a --delete "$SNAPSHOT_DIR/" "$TESTDATA/"
# Empty the snapshot dir so the EXIT trap doesn't re-restore (no-op).
rm -rf "$SNAPSHOT_DIR"
SNAPSHOT_DIR=""

echo "Step 4: Parsing captured gas values and rewriting txtar files..."
INFLATION_FACTOR=$INFLATION_FACTOR TESTDATA="$TESTDATA" OUTPUT="$OUTPUT" python3 - <<'PYEOF'
import os, re, glob

INFLATION = int(os.environ['INFLATION_FACTOR'])
TESTDATA = os.environ['TESTDATA']
OUTPUT = os.environ['OUTPUT']

def round_2sig(val):
    """Round to nearest 2 significant digits: 15658754 -> 16000000."""
    if val <= 0:
        return 100000
    s = str(val)
    n = len(s)
    if n <= 2:
        return val
    first2 = int(s[:2])
    remainder = int(s[2:])
    half = 5 * (10 ** (n - 3))
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

def is_skipped(lines, i):
    k = i - 1
    while k >= 0:
        s = lines[k].strip()
        if s == '':
            k -= 1
            continue
        if s.startswith('#'):
            if re.match(r'#\s*gas-rewrite:\s*skip\b', s):
                return True
            k -= 1
            continue
        break
    return False

# Parse test output. testscript -v prefixes each command with '> <cmd>' and
# the command's stdout/stderr lines unprefixed below it. We segment by
# command boundary so a tx that fails mid-way (prints GAS WANTED but no
# GAS USED) doesn't bind its WANTED to the next command's USED.
with open(OUTPUT, encoding='utf-8', errors='replace') as f:
    content = f.read()

current_test = None
test_gas = {}      # basename -> [(test_wanted, actual_used), ...]
pending_wanted = None

run_re      = re.compile(r'=== RUN\s+TestTestdata/(\S+)')
cmd_re      = re.compile(r'^\s*>\s+!?\s*gnokey\b')   # new gnokey command starts here
wanted_re   = re.compile(r'GAS WANTED:\s+(\d+)')
used_re     = re.compile(r'GAS USED:\s+(\d+)')
assert_re   = re.compile(r"^\s*>\s+(stdout|stderr|cmp)\b")  # assertion echoes — ignore

for line in content.split('\n'):
    m = run_re.search(line)
    if m:
        current_test = m.group(1)
        test_gas.setdefault(current_test, [])
        pending_wanted = None
        continue
    if not current_test:
        continue
    # New gnokey command boundary — discard any unmatched WANTED from prior cmd.
    if cmd_re.match(line):
        pending_wanted = None
        continue
    # Skip echoed assertion lines (testscript echoes them with '>' prefix).
    if assert_re.match(line):
        continue
    m = wanted_re.search(line)
    if m:
        pending_wanted = int(m.group(1))
    m = used_re.search(line)
    if m and pending_wanted is not None:
        test_gas[current_test].append((pending_wanted, int(m.group(1))))
        pending_wanted = None

# Walk the restored originals and apply updates.
files_updated = 0
commands_updated = 0
unmatched = []  # (basename, line_no, snippet)

for fpath in sorted(glob.glob(os.path.join(TESTDATA, '*.txtar'))):
    basename = os.path.basename(fpath).replace('.txtar', '')
    gas_pairs = test_gas.get(basename, [])

    with open(fpath, encoding='utf-8') as f:
        lines = f.readlines()

    changed = False
    gas_idx = 0

    for i, line in enumerate(lines):
        m = re.search(r'gas-wanted\s+(\d[\d_]*)', line)
        if not m:
            continue
        # Negative tests don't produce gas pairs — skip without consuming.
        if re.match(r'\s*!\s+gnokey\b', line):
            continue
        # Honor explicit opt-out marker.
        if is_skipped(lines, i):
            continue

        old_gw_str = m.group(1)
        old_gw = int(old_gw_str.replace('_', ''))
        expected_test_wanted = old_gw * INFLATION

        if gas_idx >= len(gas_pairs):
            unmatched.append((basename, i + 1, line.strip()[:80]))
            continue

        test_wanted, actual_used = gas_pairs[gas_idx]
        if test_wanted != expected_test_wanted:
            # Misalignment: the captured pair doesn't match what we expect
            # for this command. Don't consume, log, move on.
            unmatched.append((basename, i + 1, line.strip()[:80]))
            continue
        gas_idx += 1

        # Intentional-OOG guard (would only fire if 10x inflation still wasn't
        # enough — we'd want a developer to look at it).
        if actual_used > test_wanted:
            unmatched.append((basename, i + 1, f"actual {actual_used} > inflated {test_wanted}"))
            continue

        new_gw = round_2sig(int(actual_used * 1.1))

        # Idempotence band: leave alone if old_gw is in [new_gw, new_gw*2].
        # - new_gw > old_gw: must raise (current is at risk of OOG).
        # - old_gw > new_gw*2: current is wildly oversized, tighten.
        # - else: close enough; don't churn.
        if not (new_gw > old_gw or old_gw > new_gw * 2):
            continue

        new_gw_str = format_us(new_gw) if '_' in old_gw_str else str(new_gw)
        if old_gw_str.replace('_', '') == str(new_gw):
            continue

        lines[i] = line.replace(f'gas-wanted {old_gw_str}', f'gas-wanted {new_gw_str}')
        changed = True
        commands_updated += 1

        # Update the matching GAS WANTED assertion for this command. Region:
        # from this line until the next gas-wanted line (next budgeted
        # command) or a txtar file-section separator. Only rewrite assertions
        # whose value matches old_gw (other commands' assertions in the
        # region won't match and stay as-is).
        end = len(lines)
        for k in range(i + 1, len(lines)):
            if re.search(r'gas-wanted\s+\d', lines[k]) or lines[k].startswith('-- '):
                end = k
                break
        for j in range(i + 1, end):
            ga = re.search(r"GAS WANTED:\s+(\d+)", lines[j])
            if ga and ga.group(1) == str(old_gw):
                lines[j] = lines[j].replace(f'GAS WANTED: {ga.group(1)}', f'GAS WANTED: {new_gw}')

    if changed:
        with open(fpath, 'w', encoding='utf-8') as f:
            f.writelines(lines)
        files_updated += 1

print(f"  updated {commands_updated} commands across {files_updated} files")
if unmatched:
    print(f"  WARNING: {len(unmatched)} unmatched gnokey command(s):")
    for basename, lineno, snippet in unmatched[:30]:
        print(f"    {basename}.txtar:{lineno}: {snippet}")
    if len(unmatched) > 30:
        print(f"    ... and {len(unmatched) - 30} more")
PYEOF

echo "Step 5: Running UPDATE_SCRIPTS=true to refresh cmp golden files..."
INMEMORY_TS=true UPDATE_SCRIPTS=true go test -v -p 1 -timeout=30m ./gno.land/pkg/integration/ -run TestTestdata -count=1 > /dev/null 2>&1 || true

echo "Step 6: Verifying all tests pass..."
INMEMORY_TS=true go test -p 1 -timeout=30m ./gno.land/pkg/integration/ -run TestTestdata -count=1

echo "Done!"
