import { FileFlag } from 'browserfs/dist/node/core/file_flag'
import type { FileSystem } from 'browserfs/dist/node/core/file_system'
import type Stats from 'browserfs/dist/node/core/node_fs_stats'
import path from 'path-browserify'

import type { GoFileStats } from './types'

/**
 * Helper type to accept Stats without explicitly using `Stats` class func.
 * Type contains additional fields to make them writable by `normalizeStats()` and satisfy TS checks.
 */
export interface BFSStatsLike extends Stats {
  atimeMs?: number
  mtimeMs?: number
  ctimeMs?: number
}

/**
 * OpenFileFlags is Go open file flags.
 */
export const OpenFileFlags = {
  O_RDONLY: 0x0,
  O_WRONLY: 0x1,
  O_RDWR: 0x2,
  O_CREAT: 0x200,
  O_TRUNC: 0x400,
  O_APPEND: 0x8,
  O_EXCL: 0x800,
  O_SYNC: 0x80,
}

/**
 * Node.js filesystem constants used by Go.
 *
 * Valid values are required to map Go flags to Node.js flags.
 *
 * @see: https://go.dev/src/syscall/fs_js.go
 */
export const constants = {
  ...OpenFileFlags,
  UV_FS_SYMLINK_DIR: 1,
  UV_FS_SYMLINK_JUNCTION: 2,
  UV_DIRENT_UNKNOWN: 0,
  UV_DIRENT_FILE: 1,
  UV_DIRENT_DIR: 2,
  UV_DIRENT_LINK: 3,
  UV_DIRENT_FIFO: 4,
  UV_DIRENT_SOCKET: 5,
  UV_DIRENT_CHAR: 6,
  UV_DIRENT_BLOCK: 7,
  S_IFMT: 61440,
  S_IFREG: 32768,
  S_IFDIR: 16384,
  S_IFCHR: 8192,
  S_IFBLK: 24576,
  S_IFIFO: 4096,
  S_IFLNK: 40960,
  S_IFSOCK: 49152,
  UV_FS_O_FILEMAP: 0,
  O_NOCTTY: 131072,
  O_DIRECTORY: 1048576,
  O_NOFOLLOW: 256,
  O_DSYNC: 4194304,
  O_SYMLINK: 2097152,
  O_NONBLOCK: 4,
  S_IRWXU: 448,
  S_IRUSR: 256,
  S_IWUSR: 128,
  S_IXUSR: 64,
  S_IRWXG: 56,
  S_IRGRP: 32,
  S_IWGRP: 16,
  S_IXGRP: 8,
  S_IRWXO: 7,
  S_IROTH: 4,
  S_IWOTH: 2,
  S_IXOTH: 1,
  F_OK: 0,
  R_OK: 4,
  W_OK: 2,
  X_OK: 1,
  UV_FS_COPYFILE_EXCL: 1,
  COPYFILE_EXCL: 1,
  UV_FS_COPYFILE_FICLONE: 2,
  COPYFILE_FICLONE: 2,
  UV_FS_COPYFILE_FICLONE_FORCE: 4,
  COPYFILE_FICLONE_FORCE: 4,
}

export const normalizeStat = (stat: BFSStatsLike | undefined): GoFileStats | undefined => {
  if (!stat) {
    return undefined
  }

  stat.atimeMs = stat.atime.getMilliseconds()
  stat.mtimeMs = stat.mtime.getMilliseconds()
  stat.ctimeMs = stat.ctime.getMilliseconds()
  return stat as GoFileStats
}

export const normalizePath = (path: string) => {
  if (path.startsWith('./')) {
    path = path.slice(2)
  }
  if (path[0] !== '/') {
    path = '/' + path
  }
  return path
}

export const unlinkAll = async (fs: FileSystem, p: string) => {
  // Sync methods should not be called here as underlying fs might not support them.
  const isDir = (p: string) =>
    new Promise((resolve, reject) => {
      fs.stat(p, true, (err, s) => (err ? reject(err) : resolve(s?.isDirectory() ?? false)))
    })

  const rmdir = (p: string) =>
    new Promise<void>((resolve, reject) => {
      fs.rmdir(p, (err) => (err ? reject(err) : resolve()))
    })

  const readDir = (p: string) =>
    new Promise<string[]>((resolve, reject) => {
      fs.readdir(p, (err, entries) => (err ? reject(err) : resolve(entries ?? [])))
    })

  const unlink = (p: string) =>
    new Promise<void>((resolve, reject) => {
      fs.unlink(p, (err) => (err ? reject(err) : resolve()))
    })

  const delTree = async (p: string) => {
    const entryPath = normalizePath(p)
    if (await isDir(entryPath)) {
      const children = await readDir(entryPath)
      await Promise.all(children.map((name) => delTree(path.join(entryPath, name))))
      await rmdir(entryPath)
      return
    }

    await unlink(entryPath)
  }

  await delTree(p)
}

const flagsToString = (flags: number): string => {
  let result = 'r'
  if (flags & OpenFileFlags.O_WRONLY) {
    // 'w'
    result = 'w'
    if (flags & OpenFileFlags.O_EXCL) {
      result = 'wx'
    }
  } else if (flags & OpenFileFlags.O_RDWR) {
    // 'r+' or 'w+'
    if (flags & OpenFileFlags.O_CREAT || flags & OpenFileFlags.O_TRUNC) {
      // w+
      if (flags & OpenFileFlags.O_EXCL) {
        result = 'wx+'
      } else {
        result = 'w+'
      }
    } else {
      // r+
      result = 'r+'
    }
  } else if (flags & OpenFileFlags.O_APPEND) {
    result = 'a'
  }

  return result
}

/**
 * Converts Go file flags to Node.js file flags.
 * @param flags File flags number.
 */
export const convertFileFlags = (flags: number): FileFlag => {
  const flagStr = flagsToString(flags)
  return FileFlag.getFileFlag(flagStr)
}
