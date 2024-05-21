---
id: gno-tooling-gnoimports
---

# gnoimports

Gnoimports is a tool for cleaning up and organizing import statements in Gno source files. The tool automatically adds missing import statements and organizes existing ones for clearer code management.

## Features
-  **Write Mode**: Option to write organized import statements back to the source files.
-  **File Expansion**: Supports processing of single files and directories, including wildcard expansions (e.g., `path/to/dir/...`).

## Installation

To install `gnoimports`, run the following command:
```sh
make install
```
### Usage
## Basic usage:
To run gnoimports in verbose mode:
```
gnoimports -v path/to/file.gno
```

To process files and write the result back to the source file:
```
gnoimports -w path/to/file.gno
```

To process all files recursively within a directory:
```
gnoimports -v path/to/directory/...
```

### Flags

| Flag | Description                                      |
|------|--------------------------------------------------|
| -w   | Write result to (source) file instead of stdout. |
| -v   | Enable verbose mode for detailed output.         |

### Detailed Example

To recursively process all ﻿*.gno files in a directory and its subdirectories, running in verbose mode:
```
gnoimports -v path/to/directory/...
```

To process a single file and write the organized imports back to the source file:
```
gnoimports -w path/to/file.gno
```

Verbose mode example, displaying the output directly:
```
gnoimports -v path/to/file.gno
```
