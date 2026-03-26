import type { FileFlag } from 'browserfs/dist/node/core/file_flag'
import type { BFSCallback, BFSOneArgCallback } from 'browserfs/dist/node/core/file_system'

import { eReadOnly } from './utils'

/**
 * Base class to implement read-only filesystem.
 *
 * Kept separately to separate template code.
 */
export abstract class ReadOnlyFileSystem {
  /**
   * Method required to make FS consumeable by `BrowserFS.initialize()`
   */
  static isAvailable() {
    return true
  }

  diskSpace(_p: string, cb: (total: number, free: number) => any): void {
    cb(0, 0)
  }

  isReadOnly(): boolean {
    return true
  }

  supportsLinks(): boolean {
    return false
  }

  supportsProps(): boolean {
    return false
  }

  supportsSynch(): boolean {
    return true
  }

  rename(_oldPath: string, _newPath: string, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  renameSync(_oldPath: string, _newPath: string): void {
    throw eReadOnly()
  }

  unlink(_p: string, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  unlinkSync(_p: string): void {
    throw eReadOnly()
  }

  rmdir(_p: string, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  rmdirSync(_p: string): void {
    throw eReadOnly()
  }

  mkdir(_p: string, _mode: number, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  mkdirSync(_p: string, _mode: number): void {
    throw eReadOnly()
  }

  realpath(p: string, _cache: { [path: string]: string }, cb: BFSCallback<string>): void {
    // STUB
    cb(null, p)
  }

  realpathSync(p: string, _cache: { [path: string]: string }): string {
    // STUB
    return p
  }

  truncate(_p: string, _len: number, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  truncateSync(_p: string, _len: number): void {
    throw eReadOnly()
  }

  writeFile(
    _fname: string,
    _data: any,
    _encoding: string | null,
    _flag: FileFlag,
    _mode: number,
    cb: BFSOneArgCallback,
  ): void {
    cb(eReadOnly())
  }

  writeFileSync(
    _fname: string,
    _data: string | Buffer,
    _encoding: string | null,
    _flag: FileFlag,
    _mode: number,
  ): void {
    throw eReadOnly()
  }

  appendFile(
    _fname: string,
    _data: string | Buffer,
    _encoding: string | null,
    _flag: FileFlag,
    _mode: number,
    cb: BFSOneArgCallback,
  ): void {
    cb(eReadOnly())
  }

  appendFileSync(
    _fname: string,
    _data: string | Buffer,
    _encoding: string | null,
    _flag: FileFlag,
    _mode: number,
  ): void {
    throw eReadOnly()
  }

  chmod(_p: string, _isLchmod: boolean, _mode: number, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  chmodSync(_p: string, _isLchmod: boolean, _mode: number): void {
    throw eReadOnly()
  }

  chown(_p: string, _isLchown: boolean, _uid: number, _gid: number, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  chownSync(_p: string, _isLchown: boolean, _uid: number, _gid: number): void {
    throw eReadOnly()
  }

  utimes(_p: string, _atime: Date, _mtime: Date, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  utimesSync(_p: string, _atime: Date, _mtime: Date): void {
    throw eReadOnly()
  }

  link(_srcpath: string, _dstpath: string, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  linkSync(_srcpath: string, _dstpath: string): void {
    throw eReadOnly()
  }

  symlink(_srcpath: string, _dstpath: string, _type: string, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  symlinkSync(_srcpath: string, _dstpath: string, _type: string): void {
    throw eReadOnly()
  }

  readlink(p: string, cb: BFSCallback<string>): void {
    cb(null, p)
  }

  readlinkSync(p: string): string {
    return p
  }
}
