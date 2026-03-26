import type { File } from 'browserfs/dist/node/core/file'
import type { BFSOneArgCallback } from 'browserfs/dist/node/core/file_system'

/**
 * STDIN is a file descriptor for standard input.
 */
export const STDIN = 0

/**
 * STDOUT is a file descriptor for standard output.
 */
export const STDOUT = 1

/**
 * STDERR is a file descriptor for standard error.
 */
export const STDERR = 2

/**
 * FD_OFFSET is file descriptor offset.
 *
 * Starts with 3 to skip stdin, stdout and stderr.
 * First acquired file descriptor will be 3.
 */
const FD_OFFSET = STDERR + 1

/**
 * FileEntry is a file descriptor for a file.
 */
interface FileEntry {
  file: File
  path: string
}

/**
 * DirEntry is a file descriptor for a directory.
 */
interface DirEntry {
  path: string
  file?: never
}

/**
 * Generic file or directory descriptor.
 */
export type FDEntry = FileEntry | DirEntry

/**
 * FileDescriptors associates file descriptors (file IDs) with opened files.
 *
 * TODO: https://cs.opensource.google/go/go/+/master:src/syscall/tables_js.go
 */
export class FileDescriptors {
  /**
   * Pool of closed file descriptor IDs to be used for new file descriptors.
   */
  private readonly vacantFds: number[] = []

  /**
   * List of allocated file descriptors.
   */
  private readonly fds: Array<FDEntry | null> = []

  /**
   * Adds a new file descriptor entry and returns its ID.
   * Also increases global file descriptor offset.
   *
   * @param entry File descriptor entry.
   */
  add(entry: FDEntry) {
    const index = this.vacantFds.pop() ?? this.fds.length
    this.fds[index] = entry
    return index + FD_OFFSET
  }

  /**
   * Checks whether file descriptor is present in registry.
   * @param fd File descriptor ID.
   */
  get(fd: number) {
    return this.fds[fd - FD_OFFSET]
  }

  /**
   * Checks whether file descriptor is present in registry.
   * @param fd File descriptor ID.
   */
  has(fd: number) {
    return !!this.fds[fd - FD_OFFSET]
  }

  /**
   * Closes file descriptor and removes it from registry.
   * @param fd File descriptor to remove.
   */
  remove(fd: number, cb?: BFSOneArgCallback) {
    const index = fd - FD_OFFSET
    const entry = this.fds[index]
    if (!entry) {
      return
    }

    entry.file?.close((err) => cb?.(err))
    this.fds[index] = null
    this.vacantFds.push(index)
  }

  /**
   * Closes all file descriptors and removes them from registry.
   */
  dispose() {
    const closeCb: BFSOneArgCallback = (err) => {
      if (err) {
        console.error(err)
      }
    }

    this.fds.forEach((entry) => entry?.file?.close(closeCb))
    this.fds.length = 0
    this.vacantFds.length = 0
  }
}
