import { ApiError, ErrorCode } from 'browserfs/dist/node/core/api_error'
import { type File as BFSFile } from 'browserfs/dist/node/core/file'
import type { FileFlag } from 'browserfs/dist/node/core/file_flag'
import type { BFSCallback, FileSystem } from 'browserfs/dist/node/core/file_system'
import Stats, { FileType } from 'browserfs/dist/node/core/node_fs_stats'

import { abspath, PATH_SEPARATOR, trimPrefix } from '../../pathutil'
import { KVFSDir } from './dir'
import { KVFSFile } from './file'
import { ReadOnlyFileSystem } from './rofs'
import type { FileNode, FilesMap } from './types'
import { enoent, eReadOnly } from './utils'

interface KVFSOptions {
  /**
   * Whether leading root slash should be removed before checking file paths.
   *
   * Set this to true if paths in map doesn't have `/` prefix.
   */
  stripRootSlash: boolean
}

/**
 * Key-Value file system is adapter to provide files from key-value object (map).
 * Map key is full file path. Filesystem will take care about directories.
 *
 * Used to map object or store as a filesystem without copy.
 *
 * This is a bare-minimum read-only implementation.
 * Directory write operations such as mkdir are not supported.
 */
export class KeyValueFileSystem extends ReadOnlyFileSystem implements FileSystem {
  /**
   * File system name field required for BFS filesystems.
   */
  static readonly Name = 'KeyValueFileSystem'

  /**
   * Construct a filesystem instance from key-value object with files.
   *
   * @param files Map-like object of files.
   * @param stripRootSlash Whether file names in map don't start with leading slash.
   */
  constructor(
    private files: FilesMap,
    private readonly opts?: KVFSOptions,
  ) {
    super()
  }

  static fromObject(files: Record<string, FileNode>, opts?: KVFSOptions) {
    const map = new Map(Object.entries(files))
    return new KeyValueFileSystem(map, opts)
  }

  private normalizeFileKey(p: string) {
    return this.opts?.stripRootSlash ? trimPrefix(p, PATH_SEPARATOR) : p
  }

  private pathFromFileKey(key: string) {
    return this.opts?.stripRootSlash ? abspath(key) : key
  }

  private fileExists(p: string) {
    return this.files.has(this.normalizeFileKey(p))
  }

  private getFile(p: string) {
    return this.files.get(this.normalizeFileKey(p))
  }

  private dirExists(path: string) {
    if (path === PATH_SEPARATOR) {
      return true
    }

    // As original kv object is flat and doesn't have dirs - traverse all file paths.
    const pathPfx = path + PATH_SEPARATOR
    for (const key of this.files.keys()) {
      const path = this.pathFromFileKey(key)
      if (path.startsWith(pathPfx)) {
        return true
      }
    }

    return false
  }

  /**
   * Replaces referenced files object.
   * @param files
   */
  setFiles(files: FilesMap) {
    this.files = files
  }

  getName() {
    return 'kvfs'
  }

  renameSync(oldPath: string, _newPath: string): void {
    if (!this.fileExists(oldPath)) {
      throw enoent()
    }

    throw eReadOnly()
  }

  stat(p: string, isLstat: boolean | null, cb: BFSCallback<Stats>): void {
    try {
      const stats = this.statSync(p, isLstat)
      cb(null, stats)
    } catch (err) {
      cb(err as ApiError)
    }
  }

  statSync(p: string, _isLstat: boolean | null): Stats {
    const fileEnt = this.getFile(p)
    if (fileEnt) {
      return new Stats(FileType.FILE, fileEnt.content.length)
    }

    if (this.dirExists(p)) {
      return new Stats(FileType.DIRECTORY, 0)
    }

    throw enoent()
  }

  open(p: string, flag: FileFlag, mode: number, cb: BFSCallback<BFSFile>): void {
    try {
      const file = this.openSync(p, flag, mode)
      cb(null, file)
    } catch (err) {
      cb(err as ApiError)
    }
  }

  openSync(p: string, flag: FileFlag, _mode: number): BFSFile {
    if (flag.isAppendable() || flag.isWriteable()) {
      throw eReadOnly()
    }

    const fileKey = this.normalizeFileKey(p)
    const file = this.files.get(fileKey)
    if (file) {
      return new KVFSFile(file)
    }

    if (this.dirExists(p)) {
      return new KVFSDir()
    }

    throw enoent()
  }

  readdir(p: string, cb: BFSCallback<string[]>): void {
    try {
      const result = this.readdirSync(p)
      cb(null, result)
    } catch (err) {
      cb(err as ApiError)
    }
  }

  readdirSync(p: string): string[] {
    const pathPfx = p === PATH_SEPARATOR ? p : p + PATH_SEPARATOR
    const children = new Set<string>()
    for (const key of this.files.keys()) {
      const path = this.pathFromFileKey(key)
      if (path === p) {
        throw new ApiError(ErrorCode.ENOTDIR)
      }

      if (!path.startsWith(pathPfx)) {
        continue
      }

      // Keep only first element of path in result to respect deep paths.
      const relRoot = path.slice(pathPfx.length)
      const slashIndex = relRoot.indexOf(PATH_SEPARATOR)
      children.add(slashIndex !== -1 ? relRoot.slice(0, slashIndex) : relRoot)
    }

    if (!children.size) {
      throw enoent()
    }

    return Array.from(children)
  }

  exists(p: string, cb: (exists: boolean) => void): void {
    cb(this.existsSync(p))
  }

  existsSync(p: string): boolean {
    return this.fileExists(p) || this.dirExists(p)
  }

  readFile(fname: string, encoding: string | null, flag: FileFlag, cb: BFSCallback<string | Buffer>): void {
    try {
      const result = this.readFileSync(fname, encoding, flag)
      cb(null, result)
    } catch (err) {
      cb(err as ApiError)
    }
  }

  readFileSync(fname: string, _encoding: string | null, _flag: FileFlag): string {
    const file = this.getFile(fname)
    if (file) {
      return file.content
    }

    throw this.dirExists(fname) ? new ApiError(ErrorCode.EISDIR) : enoent()
  }
}
