# Historical-tx patches for test-13

Every `*.jsonl` in this directory is picked up by `apply-test-13-replay.sh`
and passed to `gnogenesis fork generate` as `--patch-txs <path>`. The format
is `AnnotatedTx` (one entry per line):

```jsonl
{"reason":"why we patched this tx","tx":{...},"metadata":{...}}
```

## Authoring a patch

1. Find the original tx in the source stream — e.g. against the cached
   `work/phase-2/txs.jsonl`:

   ```bash
   jq -c 'select(.metadata.block_height == 1950)' work/phase-2/txs.jsonl
   ```

2. Copy that line into a new `.jsonl` in this directory, prepend a
   `"reason": "..."` field, and edit `tx.msg[0].package.files[0].body` to the
   post-#5669 (or whatever current master demands) gno source.

3. Re-run phase-2; `gnogenesis fork generate` matches the patch against the
   source tx by `(block_height, signer_info[0].address, signer_info[0].sequence)`
   and substitutes the body. Original tx + reason are preserved on
   `metadata.{original_tx, note, source="patched"}` for the inspect report.

## Validation

`gnogenesis fork generate` fails fast on:
- Any patch key matching no source tx (typo guard)
- Same key in two different patch files (conflict)
- Same key twice in the same patch file (typo guard)

## Files

One file per concern. Suggested naming:
- `unrestrict.jsonl` — patches for `unrestrict.gno` MsgRuns at various heights
- `restrict.jsonl` — patch for `restrict.gno` MsgRun
- `set-cla.jsonl` — patch for `set_cla.gno` MsgRun
- `set-minfee.jsonl` — patch for `set_minfee.gno` MsgRun
- `v2-valset-noops.jsonl` — empty-body noops for `add_validator.gno` /
  `rm_validator.gno` MsgRuns targeting the vestigial r/sys/validators/v2 realm
