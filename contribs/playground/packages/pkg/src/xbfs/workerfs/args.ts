/* eslint-disable @typescript-eslint/no-unsafe-function-type */
/**
 * WorkerFS implementation copy from BrowserFS package.
 *
 * Source: browserfs@2.0.0/src/core/file_flag.ts
 *
 * @see `index.ts` for background.
 */

import { ApiError } from 'browserfs/dist/node/core/api_error'
import { type File } from 'browserfs/dist/node/core/file'
import { FileFlag } from 'browserfs/dist/node/core/file_flag'
import { type BFSCallback, type BFSOneArgCallback } from 'browserfs/dist/node/core/file_system'
import global from 'browserfs/dist/node/core/global'
import Stats from 'browserfs/dist/node/core/node_fs_stats'
import { arrayBuffer2Buffer, buffer2ArrayBuffer } from 'browserfs/dist/node/core/util'

export interface IBrowserFSMessage {
  browserfsMessage: boolean
}

export enum SpecialArgType {
  // Callback
  CB,
  // File descriptor
  FD,
  // API error
  API_ERROR,
  // Stats object
  STATS,
  // Initial probe for file system information.
  PROBE,
  // FileFlag object.
  FILEFLAG,
  // Buffer object.
  BUFFER,
  // Generic Error object.
  ERROR,
}

export interface IAPIErrorArgument extends ISpecialArgument {
  // The error object, as an array buffer.
  errorData: ArrayBuffer | SharedArrayBuffer
}

export function apiErrorLocal2Remote(e: ApiError): IAPIErrorArgument {
  return {
    type: SpecialArgType.API_ERROR,
    errorData: bufferToTransferrableObject(e.writeToBuffer()),
  }
}

export function apiErrorRemote2Local(e: IAPIErrorArgument): ApiError {
  return ApiError.fromBuffer(transferrableObjectToBuffer(e.errorData))
}

export interface IErrorArgument extends ISpecialArgument {
  // The name of the error (e.g. 'TypeError').
  name: string
  // The message associated with the error.
  message: string
  // The stack associated with the error.
  stack: string
}

export function errorLocal2Remote(e: Error): IErrorArgument {
  return {
    type: SpecialArgType.ERROR,
    name: e.name,
    message: e.message,
    stack: e.stack!,
  }
}

export function errorRemote2Local(e: IErrorArgument): Error {
  let ctor: new (msg: string) => Error = global[e.name]
  if (typeof ctor !== 'function') {
    ctor = Error
  }

  const err = new ctor(e.message)
  err.stack = e.stack
  return err
}

export interface IStatsArgument extends ISpecialArgument {
  // The stats object as an array buffer.
  statsData: ArrayBuffer | SharedArrayBuffer
}

export function statsLocal2Remote(stats: Stats): IStatsArgument {
  return {
    type: SpecialArgType.STATS,
    statsData: bufferToTransferrableObject(stats.toBuffer()),
  }
}

export function statsRemote2Local(stats: IStatsArgument): Stats {
  return Stats.fromBuffer(transferrableObjectToBuffer(stats.statsData))
}

export interface IFileFlagArgument extends ISpecialArgument {
  flagStr: string
}

export function fileFlagLocal2Remote(flag: FileFlag): IFileFlagArgument {
  return {
    type: SpecialArgType.FILEFLAG,
    flagStr: flag.getFlagString(),
  }
}

export function fileFlagRemote2Local(remoteFlag: IFileFlagArgument): FileFlag {
  return FileFlag.getFileFlag(remoteFlag.flagStr)
}

export interface IBufferArgument extends ISpecialArgument {
  data: ArrayBuffer | SharedArrayBuffer
}

export function bufferToTransferrableObject(buff: Buffer): ArrayBuffer | SharedArrayBuffer {
  return buffer2ArrayBuffer(buff)
}

export function transferrableObjectToBuffer(buff: ArrayBuffer | SharedArrayBuffer): Buffer {
  return arrayBuffer2Buffer(buff)
}

export function bufferLocal2Remote(buff: Buffer): IBufferArgument {
  return {
    type: SpecialArgType.BUFFER,
    data: bufferToTransferrableObject(buff),
  }
}

export function bufferRemote2Local(buffArg: IBufferArgument): Buffer {
  return transferrableObjectToBuffer(buffArg.data)
}

export interface IAPIRequest extends IBrowserFSMessage {
  method: string
  args: Array<number | string | ISpecialArgument>
}

export function isAPIRequest(data: any): data is IAPIRequest {
  return data && typeof data === 'object' && data.hasOwnProperty('browserfsMessage') && data.browserfsMessage
}

export interface IAPIResponse extends IBrowserFSMessage {
  cbId: number
  args: Array<number | string | ISpecialArgument>
}

export function isAPIResponse(data: any): data is IAPIResponse {
  return data && typeof data === 'object' && data.hasOwnProperty('browserfsMessage') && data.browserfsMessage
}

export interface ISpecialArgument {
  type: SpecialArgType
}

export interface IProbeResponse extends ISpecialArgument {
  isReadOnly: boolean
  supportsLinks: boolean
  supportsProps: boolean
}

export interface ICallbackArgument extends ISpecialArgument {
  // The callback ID.
  id: number
}

export class CallbackArgumentConverter {
  private _callbacks: { [id: number]: Function } = {}
  private _nextId = 0

  public toRemoteArg(cb: Function): ICallbackArgument {
    const id = this._nextId++
    this._callbacks[id] = cb
    return {
      type: SpecialArgType.CB,
      id,
    }
  }

  public toLocalArg(id: number): Function {
    const cb = this._callbacks[id]
    delete this._callbacks[id]
    return cb
  }
}

export interface IFileDescriptorArgument extends ISpecialArgument {
  // The file descriptor's id on the remote side.
  id: number
  // The entire file's data, as an array buffer.
  data: ArrayBuffer | SharedArrayBuffer
  // The file's stat object, as an array buffer.
  stat: ArrayBuffer | SharedArrayBuffer
  // The path to the file.
  path: string
  // The flag of the open file descriptor.
  flag: string
}

export class FileDescriptorArgumentConverter {
  private _fileDescriptors: { [id: number]: File } = {}
  private _nextId = 0

  public toRemoteArg(fd: File, p: string, flag: FileFlag, cb: BFSCallback<IFileDescriptorArgument>): void {
    const id = this._nextId++
    let data: ArrayBuffer | SharedArrayBuffer
    let stat: ArrayBuffer | SharedArrayBuffer
    this._fileDescriptors[id] = fd

    // Extract needed information asynchronously.
    fd.stat((err, stats) => {
      if (err) {
        cb(err)
      } else {
        stat = bufferToTransferrableObject(stats!.toBuffer())
        // If it's a readable flag, we need to grab contents.
        if (flag.isReadable()) {
          fd.read(
            Buffer.alloc(stats!.size),
            0,
            stats!.size,
            0,
            (err?: ApiError | null, _bytesRead?: number, buff?: Buffer) => {
              if (err) {
                cb(err)
              } else {
                data = bufferToTransferrableObject(buff!)
                cb(null, {
                  type: SpecialArgType.FD,
                  id,
                  data,
                  stat,
                  path: p,
                  flag: flag.getFlagString(),
                })
              }
            },
          )
        } else {
          // File is not readable, which means writing to it will append or
          // truncate/replace existing contents. Return an empty arraybuffer.
          cb(null, {
            type: SpecialArgType.FD,
            id,
            data: new ArrayBuffer(0),
            stat,
            path: p,
            flag: flag.getFlagString(),
          })
        }
      }
    })
  }

  public applyFdAPIRequest(request: IAPIRequest, cb: BFSOneArgCallback): void {
    const fdArg = request.args[0] as IFileDescriptorArgument
    this._applyFdChanges(fdArg, (err, fd?) => {
      if (err) {
        cb(err)
      } else {
        // Apply method on now-changed file descriptor.
        // @ts-ignore
        fd?.[request.method]((e?: ApiError) => {
          if (request.method === 'close') {
            delete this._fileDescriptors[fdArg.id]
          }
          cb(e)
        })
      }
    })
  }

  private _applyFdChanges(remoteFd: IFileDescriptorArgument, cb: BFSCallback<File>): void {
    const fd = this._fileDescriptors[remoteFd.id]
    const data = transferrableObjectToBuffer(remoteFd.data)
    const remoteStats = Stats.fromBuffer(transferrableObjectToBuffer(remoteFd.stat))

    // Write data if the file is writable.
    const flag = FileFlag.getFileFlag(remoteFd.flag)
    if (flag.isWriteable()) {
      // Appendable: Write to end of file.
      // Writeable: Replace entire contents of file.
      fd.write(data, 0, data.length, flag.isAppendable() ? fd.getPos()! : 0, (e?: ApiError | null) => {
        function applyStatChanges() {
          // Check if mode changed.
          fd.stat((e, stats?) => {
            if (e) {
              cb(e)
            } else {
              if (stats!.mode !== remoteStats.mode) {
                fd.chmod(remoteStats.mode, (e: any) => {
                  cb(e, fd)
                })
              } else {
                cb(e, fd)
              }
            }
          })
        }
        if (e) {
          cb(e)
        } else {
          // If writeable & not appendable, we need to ensure file contents are
          // identical to those from the remote FD. Thus, we truncate to the
          // length of the remote file.
          if (!flag.isAppendable()) {
            fd.truncate(data.length, () => {
              applyStatChanges()
            })
          } else {
            applyStatChanges()
          }
        }
      })
    } else {
      cb(null, fd)
    }
  }
}
