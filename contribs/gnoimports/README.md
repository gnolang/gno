# GNOIMPORTS

## NAME
gnoimports - Run Gno imports cleanup

## SYNOPSIS
**gnoimports** [*flags*] [*path* ...]

## DESCRIPTION
The **gnoimports** command processes and cleans up Gno source files by specifying paths and using various flags for output control.

## OPTIONS
**-w**  
  Write result to (source) file instead of stdout.

**-v**  
  Enable verbose mode.

## USAGE
### Basic usage:

To process files and write the result back:

```
gnoimports -w path/to/file.gno
```

To process all files recursively within a directory:
```
gnoimports -v path/to/directory/...
```

