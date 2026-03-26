import { BackendInmemory } from '@gnostudio/pkg'

import * as BrowserFS from 'browserfs'
import type { FileSystem } from 'browserfs/dist/node/core/file_system'

import { SyscallError } from './common/errors'
import { FileDescriptors, STDERR, STDIN, STDOUT } from './common/fd'
import {
  wrapErrCallback,
  wrapResultCallback,
  type ErrCallback,
  type ErrorType,
  type GoFileStats,
  type ResultCallback,
  type ThreeArgCallback,
} from './common/types'
import { constants, convertFileFlags, normalizePath, normalizeStat, type BFSStatsLike } from './common/utils'
import { newLogOutputWriter, newNoopInputHandler, type InputHandler, type OutputHandler } from './stdio'

const { Buffer } = BrowserFS.BFSRequire('buffer')

export interface StdioHandlers {
  stdin: InputHandler
  stdout: OutputHandler
  stderr: OutputHandler
}

/**
 * A TypeScript helper to extract a type from a Promise type.
 * Used for WasmFS type declaration.
 */
type UnwrapPromise<T> = T extends Promise<infer U> ? U : T

/**
 * WasmFS is a Go-compatible filesystem.
 */
export type WasmFS = UnwrapPromise<ReturnType<typeof createWasmFS>>

/**
 * Fills defaults for missing stdio handlers.
 * @param src
 * @returns
 */
const prefillDefaultHandlers = (src?: Partial<StdioHandlers>): StdioHandlers => {
  const stdin = src?.stdin ?? newNoopInputHandler()
  const stdout = src?.stdout ?? newLogOutputWriter()
  const stderr = src?.stderr ?? newLogOutputWriter(true)

  return { stdin, stdout, stderr }
}

/**
 * Creates Go-compatible filesysten from BrowserFS.
 *
 * @param fs BrowserFS instance
 * @param stdio Stdio handlers. If not provided, default console logger will be used as output.
 */
export const createWasmFS = async (baseFs?: FileSystem | null, stdio?: Partial<StdioHandlers>) => {
  const fs = baseFs ?? BrowserFS.initialize(await BackendInmemory())
  const fds = new FileDescriptors()

  const { stdin, stdout, stderr } = prefillDefaultHandlers(stdio)

  const wasmfs = {
    constants,
    writeSync: (fd: number, buf: Uint8Array, offset: number, length: number, position: number): number => {
      switch (fd) {
        case STDOUT: {
          stdout.write(buf.slice(0))
          return buf.length
        }

        case STDERR:
          stderr.write(buf.slice(0))
          return buf.length

        default: {
          // This method isn't used by Go and should not be used by anybody as provided fs is async.
          // Kept for compatibility with Go's FS interface.
          console.warn('wasmfs: writeSync method usage is dangerous and not supported on remote file systems')
          const fdEntry = fds.get(fd)
          if (!fdEntry) {
            throw SyscallError.EBADF()
          }

          if (!fdEntry.file) {
            throw SyscallError.EISDIR()
          }

          const n = fdEntry.file.writeSync(Buffer.from(buf), offset, length, position)
          return n
        }
      }
    },
    write: (
      fd: number,
      buf: Uint8Array,
      offset: number,
      length: number,
      position: number,
      cb: ResultCallback<number>,
    ): void => {
      switch (fd) {
        case STDOUT: {
          stdout.write(buf.slice(0))
          cb(null, buf.length)
          return
        }

        case STDERR:
          stderr.write(buf.slice(0))
          cb(null, buf.length)
          return

        default: {
          const fdEntry = fds.get(fd)
          if (!fdEntry) {
            cb(SyscallError.EBADF())
            return
          }

          if (!fdEntry.file) {
            cb(SyscallError.EISDIR())
            return
          }

          fdEntry.file.write(Buffer.from(buf), offset, length, position, wrapResultCallback(cb))
        }
      }
    },
    read: (
      fd: number,
      buf: Uint8Array | Buffer,
      offset: number,
      length: number,
      position: number,
      callback: ThreeArgCallback<number, Uint8Array>,
    ): void => {
      if (fd === STDIN) {
        stdin.read(buf as unknown as Uint8Array, callback)
        return
      }

      const fdEntry = fds.get(fd)
      if (!fdEntry) {
        callback(SyscallError.EBADF())
        return
      }

      if (!fdEntry.file) {
        callback(SyscallError.EISDIR())
        return
      }

      // Some BFS filesystems such as WorkerFS strictly require buffer to be a "Buffer" instance.
      const nodeBuff = Buffer.isBuffer(buf) ? buf : Buffer.from(buf.buffer)
      const cb = (err?: ErrorType, n?: number) =>
        callback(SyscallError.normalizeError(err), n, nodeBuff as unknown as Uint8Array)
      fdEntry.file.read(nodeBuff, offset, length, position, cb)
    },
    open: (path: string, flags: number, mode: number, cb: ResultCallback<number>): void => {
      const nodeFlags = convertFileFlags(flags)
      if (nodeFlags.isAppendable()) {
        cb(SyscallError.ENOSYS())
        return
      }

      const fpath = normalizePath(path)
      fs.open(fpath, nodeFlags, mode, (err, file) => {
        if (err?.code === 'EISDIR') {
          // Gno does fopen() on dirs before readdir(). Permit fopen on dirs.
          const fd = fds.add({ path: fpath })
          cb(null, fd)
          return
        }

        if (err) {
          cb(SyscallError.normalizeError(err))
          return
        }

        const fd = fds.add({ file, path: fpath })
        cb(null, fd)
      })
    },
    chmod: (_path: string, _mode: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    chown: (_path: string, _uid: number, _gid: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    close: (fd: number, callback: ErrCallback): void => {
      if (!fds.has(fd)) {
        callback(SyscallError.EBADF())
        return
      }

      fds.remove(fd)
      callback(null)
    },
    fchmod: (_fd: number, _mode: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    fchown: (_fd: number, _uid: number, _gid: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    fstat: (fd: number, callback: ResultCallback<GoFileStats>): void => {
      const f = fds.get(fd)
      if (!f) {
        callback(SyscallError.EBADF())
        return
      }

      const cb = (err?: ErrorType | null, stats?: BFSStatsLike) =>
        callback(SyscallError.normalizeError(err), normalizeStat(stats))

      f.file ? f.file.stat(cb) : fs.stat(normalizePath(f.path), false, cb)
    },
    fsync: (_fd: number, callback: ErrCallback): void => {
      callback(null)
    },
    ftruncate: (fd: number, length: number, callback: ErrCallback): void => {
      const entry = fds.get(fd)
      if (!entry) {
        callback(SyscallError.EBADF())
        return
      }

      // Deny ftruncate on directories.
      if (!entry.file) {
        callback(SyscallError.EISDIR())
        return
      }

      entry.file.truncate(length, wrapErrCallback(callback))
    },
    lchown: (_path: string, _uid: number, _gid: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    link: (_path: string, _link: string, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    lstat: (path: string, callback: ResultCallback<GoFileStats>): void => {
      fs.stat(normalizePath(path), true, (err, stat) => callback(SyscallError.normalizeError(err), normalizeStat(stat)))
    },
    mkdir: (path: string, perm: number, callback: ErrCallback): void => {
      fs.mkdir(path, perm, wrapErrCallback(callback))
    },
    readdir: (path: string, callback: ResultCallback<string[]>): void => {
      fs.readdir(normalizePath(path), wrapResultCallback(callback))
    },
    readlink: (_path: string, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    rename: (_from: string, _to: string, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    rmdir: (path: string, callback: ErrCallback): void => {
      fs.rmdir(path, wrapErrCallback(callback))
    },
    stat: (path: string, callback: ResultCallback<GoFileStats>): void => {
      wasmfs.lstat(path, wrapResultCallback(callback))
    },
    symlink: (_path: string, _link: string, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    truncate: (_path: string, _length: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    unlink: (path: string, callback: ErrCallback): void => {
      fs.unlink(path, wrapErrCallback(callback))
    },
    utimes: (_path: string, _atime: number, _mtime: number, callback: ErrCallback): void => {
      callback(SyscallError.ENOSYS())
    },
    onExit: (): void => {
      fds.dispose()
      stderr.close?.()
      stdout.close?.()
      stdin.close?.()
    },
  }

  return wasmfs
}
