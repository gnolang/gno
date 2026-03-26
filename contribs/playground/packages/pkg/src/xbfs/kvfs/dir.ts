import { ErrorCode } from 'browserfs/dist/node/core/api_error'
import { BaseFile, type File as BFSFile } from 'browserfs/dist/node/core/file'
import type { BFSCallback, BFSOneArgCallback, BFSThreeArgCallback } from 'browserfs/dist/node/core/file_system'
import Stats, { FileType } from 'browserfs/dist/node/core/node_fs_stats'

import { newApiError } from './utils'

/**
 * Stub class for directories to implement BFS file interface.
 *
 * Necessary as directories can be fopen'ed before reading by Go.
 */
export class KVFSDir extends BaseFile implements BFSFile {
  getPos(): number | undefined {
    return undefined
  }

  stat(cb: BFSCallback<Stats>): void {
    cb(null, new Stats(FileType.DIRECTORY, 0))
  }

  statSync(): Stats {
    return new Stats(FileType.DIRECTORY, 0)
  }

  close(cb: BFSOneArgCallback): void {
    cb()
  }

  closeSync(): void {
    // noop
  }

  truncate(_len: number, cb: BFSOneArgCallback): void {
    cb(newApiError(ErrorCode.EISDIR))
  }

  truncateSync(_len: number): void {
    throw newApiError(ErrorCode.EISDIR)
  }

  write(
    _buffer: Buffer,
    _offset: number,
    _length: number,
    _position: number | null,
    cb: BFSThreeArgCallback<number, Buffer>,
  ): void {
    cb(newApiError(ErrorCode.EISDIR))
  }

  writeSync(_buffer: Buffer, _offset: number, _length: number, _position: number | null): number {
    throw newApiError(ErrorCode.EISDIR)
  }

  read(
    _buffer: Buffer,
    _offset: number,
    _length: number,
    _position: number | null,
    cb: BFSThreeArgCallback<number, Buffer>,
  ): void {
    cb(newApiError(ErrorCode.EISDIR))
  }

  readSync(_buffer: Buffer, _offset: number, _length: number, _position: number): number {
    throw newApiError(ErrorCode.EISDIR)
  }
}
