import * as BrowserFS from 'browserfs'
import { ErrorCode, type ApiError } from 'browserfs/dist/node/core/api_error'
import { BaseFile, type File as BFSFile } from 'browserfs/dist/node/core/file'
import type { BFSCallback, BFSOneArgCallback, BFSThreeArgCallback } from 'browserfs/dist/node/core/file_system'
import Stats, { FileType } from 'browserfs/dist/node/core/node_fs_stats'

import type { FileNode } from './types'
import { eReadOnly, newApiError } from './utils'

const { Buffer } = BrowserFS.BFSRequire('buffer')

/**
 * File interface implementation to implement file reading from object.
 */
export class KVFSFile extends BaseFile implements BFSFile {
  private offset = 0
  private isClosed = false
  constructor(private readonly entry: FileNode) {
    super()
  }

  getPos(): number | undefined {
    return this.offset
  }

  stat(cb: BFSCallback<Stats>): void {
    try {
      cb(null, this.statSync())
    } catch (err) {
      cb(err as ApiError)
    }
  }

  statSync(): Stats {
    if (this.isClosed) {
      throw newApiError(ErrorCode.EBADF)
    }

    return new Stats(FileType.FILE, this.entry.content.length)
  }

  close(cb: BFSOneArgCallback): void {
    this.isClosed = true
    cb(null)
  }

  closeSync(): void {
    this.isClosed = true
  }

  truncate(_len: number, cb: BFSOneArgCallback): void {
    cb(eReadOnly())
  }

  truncateSync(_len: number): void {
    throw eReadOnly()
  }

  write(
    _buffer: Buffer,
    _offset: number,
    _length: number,
    _position: number | null,
    cb: BFSThreeArgCallback<number, Buffer>,
  ): void {
    cb(eReadOnly())
  }

  writeSync(_buffer: Buffer, _offset: number, _length: number, _position: number | null): number {
    throw eReadOnly()
  }

  read(
    buffer: Buffer,
    offset: number,
    length: number,
    position: number | null,
    cb: BFSThreeArgCallback<number, Buffer>,
  ): void {
    try {
      const n = this.readSync(buffer, offset, length, position ?? this.offset)
      cb(null, n, buffer)
    } catch (err) {
      cb(err as ApiError)
    }
  }

  readSync(buffer: Buffer, offset: number, length: number, position: number): number {
    const contentLength = this.entry.content.length

    // position might be null although interface doesn't tell this.
    position = position ?? this.offset

    // If the position is beyond the end of the string, return 0 because no data can be read.
    if (position >= contentLength) {
      return 0
    }

    // Determine the actual number of characters to read
    const endPosition = Math.min(position + length, contentLength)
    const substring = this.entry.content.substring(position, endPosition)

    // Convert the substring into a buffer
    const tempBuffer = Buffer.from(substring, 'utf-8')

    // Copy the data from tempBuffer to the provided buffer at the given offset
    const bytesWritten = tempBuffer.copy(buffer as unknown as Uint8Array<ArrayBufferLike>, offset)

    if (position === this.offset) {
      this.offset += bytesWritten
    }

    return bytesWritten
  }
}
