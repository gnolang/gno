/* eslint-disable */
/**
 * WorkerFS implementation copy from BrowserFS package.
 *
 * Source: browserfs@2.0.0/src/core/file_flag.ts
 *
 * @see `index.ts` for background.
 */

import { ApiError, ErrorCode } from 'browserfs/dist/node/core/api_error'
import { BaseFile, type File } from 'browserfs/dist/node/core/file'
import { FileFlag } from 'browserfs/dist/node/core/file_flag'
import {
  BaseFileSystem,
  type BFSCallback,
  type BFSOneArgCallback,
  type FileSystem,
  type FileSystemOptions,
} from 'browserfs/dist/node/core/file_system'
import Stats from 'browserfs/dist/node/core/node_fs_stats'
import { emptyBuffer } from 'browserfs/dist/node/core/util'
import PreloadFile from 'browserfs/dist/node/generic/preload_file'

import {
  apiErrorLocal2Remote,
  apiErrorRemote2Local,
  bufferLocal2Remote,
  bufferRemote2Local,
  bufferToTransferrableObject,
  CallbackArgumentConverter,
  errorLocal2Remote,
  errorRemote2Local,
  FileDescriptorArgumentConverter,
  fileFlagLocal2Remote,
  fileFlagRemote2Local,
  isAPIRequest,
  isAPIResponse,
  SpecialArgType,
  statsLocal2Remote,
  statsRemote2Local,
  transferrableObjectToBuffer,
  type IAPIErrorArgument,
  type IAPIRequest,
  type IAPIResponse,
  type IBufferArgument,
  type ICallbackArgument,
  type IErrorArgument,
  type IFileDescriptorArgument,
  type IFileFlagArgument,
  type IProbeResponse,
  type ISpecialArgument,
  type IStatsArgument,
} from './args'

export interface WorkerFSOptions {
  worker: MessagePort | Worker
}

/**
 * Represents a remote file in a different worker/thread.
 */
class WorkerFile extends PreloadFile<WorkerFS> {
  private readonly _remoteFdId: number

  constructor(_fs: WorkerFS, _path: string, _flag: FileFlag, _stat: Stats, remoteFdId: number, contents?: Buffer) {
    super(_fs, _path, _flag, _stat, contents)
    this._remoteFdId = remoteFdId
  }

  public getRemoteFdId() {
    return this._remoteFdId
  }

  /**
   * @hidden
   */
  public toRemoteArg(): IFileDescriptorArgument {
    return {
      type: SpecialArgType.FD,
      id: this._remoteFdId,
      data: bufferToTransferrableObject(this.getBuffer()),
      stat: bufferToTransferrableObject(this.getStats().toBuffer()),
      path: this.getPath(),
      flag: this.getFlag().getFlagString(),
    }
  }

  public sync(cb: BFSOneArgCallback): void {
    this._syncClose('sync', cb)
  }

  public close(cb: BFSOneArgCallback): void {
    this._syncClose('close', cb)
  }

  private _syncClose(type: string, cb: BFSOneArgCallback): void {
    if (this.isDirty()) {
      // @ts-expect-error -- keep original code.
      this._fs.syncClose(type, this, (e?: ApiError) => {
        if (!e) {
          this.resetDirty()
        }
        cb(e)
      })
    } else {
      cb()
    }
  }
}

export class WorkerFS extends BaseFileSystem implements FileSystem {
  public static readonly Name = 'WorkerFS'

  public static readonly Options: FileSystemOptions = {
    worker: {
      type: 'object',
      description: 'The target worker that you want to connect to, or the current worker if in a worker context.',
      validator: function (v: object, cb: BFSOneArgCallback): void {
        // Check for a `postMessage` function.
        if ('postMessage' in v) {
          cb()
        } else {
          cb(new ApiError(ErrorCode.EINVAL, `option must be a Worker or MessagePort instance.`))
        }
      },
    },
  }

  public static Create(opts: WorkerFSOptions, cb: BFSCallback<WorkerFS>): void {
    const fs = new WorkerFS(opts.worker)
    fs._initialize(() => {
      cb(null, fs)
    })
  }

  public static isAvailable(): boolean {
    // CHANGE: we don't need that.
    // return typeof importScripts !== 'undefined' || typeof Worker !== 'undefined'
    return true
  }

  /**
   * Attaches a listener to the remote worker for file system requests.
   */
  public static attachRemoteListener(rootFs: FileSystem, port: MessagePort | Worker) {
    const fdConverter = new FileDescriptorArgumentConverter()

    function argLocal2Remote(arg: any, requestArgs: any[], cb: BFSCallback<any>): void {
      switch (typeof arg) {
        case 'object':
          if (arg instanceof Stats) {
            cb(null, statsLocal2Remote(arg))
          } else if (arg instanceof ApiError) {
            cb(null, apiErrorLocal2Remote(arg))
          } else if (arg instanceof BaseFile) {
            // Pass in p and flags from original request.
            cb(null, fdConverter.toRemoteArg(arg as File, requestArgs[0], requestArgs[1], cb))
          } else if (arg instanceof FileFlag) {
            cb(null, fileFlagLocal2Remote(arg))
          } else if (arg instanceof Buffer) {
            cb(null, bufferLocal2Remote(arg))
          } else if (arg instanceof Error) {
            cb(null, errorLocal2Remote(arg))
          } else {
            cb(null, arg)
          }
          break
        default:
          cb(null, arg)
          break
      }
    }

    function argRemote2Local(arg: any, fixedRequestArgs: any[]): any {
      if (!arg) {
        return arg
      }
      switch (typeof arg) {
        case 'object':
          if (typeof arg.type === 'number') {
            const specialArg = arg as ISpecialArgument
            switch (specialArg.type) {
              case SpecialArgType.CB: {
                const cbId = (arg as ICallbackArgument).id
                return function () {
                  let i: number
                  const fixedArgs = new Array(arguments.length)
                  let message: IAPIResponse
                  let countdown = arguments.length

                  function abortAndSendError(err: ApiError) {
                    if (countdown > 0) {
                      countdown = -1
                      message = {
                        browserfsMessage: true,
                        cbId,
                        args: [apiErrorLocal2Remote(err)],
                      }
                      port.postMessage(message)
                    }
                  }

                  for (i = 0; i < arguments.length; i++) {
                    // Capture i and argument.
                    ;((i: number, arg: any) => {
                      argLocal2Remote(arg, fixedRequestArgs, (err, fixedArg?) => {
                        fixedArgs[i] = fixedArg
                        if (err) {
                          abortAndSendError(err)
                        } else if (--countdown === 0) {
                          message = {
                            browserfsMessage: true,
                            cbId,
                            args: fixedArgs,
                          }
                          port.postMessage(message)
                        }
                      })
                    })(i, arguments[i])
                  }

                  if (arguments.length === 0) {
                    message = {
                      browserfsMessage: true,
                      cbId,
                      args: fixedArgs,
                    }
                    port.postMessage(message)
                  }
                }
              }
              case SpecialArgType.API_ERROR:
                return apiErrorRemote2Local(specialArg as IAPIErrorArgument)
              case SpecialArgType.STATS:
                return statsRemote2Local(specialArg as IStatsArgument)
              case SpecialArgType.FILEFLAG:
                return fileFlagRemote2Local(specialArg as IFileFlagArgument)
              case SpecialArgType.BUFFER:
                return bufferRemote2Local(specialArg as IBufferArgument)
              case SpecialArgType.ERROR:
                return errorRemote2Local(specialArg as IErrorArgument)
              default:
                // No idea what this is.
                return arg
            }
          } else {
            return arg
          }
        default:
          return arg
      }
    }

    port.onmessage = (e: MessageEvent) => {
      const request: object = e.data
      if (isAPIRequest(request)) {
        const args = request.args
        const fixedArgs = new Array<any>(args.length)

        switch (request.method) {
          case 'close':
          case 'sync':
            ;(() => {
              // File descriptor-relative methods.
              const remoteCb = args[1] as ICallbackArgument
              fdConverter.applyFdAPIRequest(request, (err?: ApiError | null) => {
                // Send response.
                const response: IAPIResponse = {
                  browserfsMessage: true,
                  cbId: remoteCb.id,
                  args: err ? [apiErrorLocal2Remote(err)] : [],
                }
                port.postMessage(response)
              })
            })()
            break
          case 'probe':
            ;(() => {
              const remoteCb = args[1] as ICallbackArgument
              const probeResponse: IProbeResponse = {
                type: SpecialArgType.PROBE,
                isReadOnly: rootFs.isReadOnly(),
                supportsLinks: rootFs.supportsLinks(),
                supportsProps: rootFs.supportsProps(),
              }
              const response: IAPIResponse = {
                browserfsMessage: true,
                cbId: remoteCb.id,
                args: [probeResponse],
              }

              port.postMessage(response)
            })()
            break
          default: {
            // File system methods.
            for (let i = 0; i < args.length; i++) {
              fixedArgs[i] = argRemote2Local(args[i], fixedArgs)
            }

            // TODO: don't use global fs instance.
            ;((rootFs as any)[request.method] as Function).apply(rootFs, fixedArgs)
            break
          }
        }
      }
    }
  }

  private readonly _worker: MessagePort | Worker
  private readonly _callbackConverter = new CallbackArgumentConverter()

  private _isInitialized = false
  private _isReadOnly = false
  private _supportLinks = false
  private _supportProps = false

  /**
   * Constructs a new WorkerFS instance that connects with BrowserFS running on
   * the specified worker.
   */
  private constructor(target: MessagePort | Worker) {
    super()
    this._worker = target
    this._worker.onmessage = (e: MessageEvent) => {
      const resp: object = e.data
      if (isAPIResponse(resp)) {
        let i: number
        const args = resp.args
        const fixedArgs = new Array(args.length)
        // Dispatch event to correct id.
        for (i = 0; i < fixedArgs.length; i++) {
          fixedArgs[i] = this._argRemote2Local(args[i])
        }
        this._callbackConverter.toLocalArg(resp.cbId).apply(null, fixedArgs)
      }
    }
  }

  public getName(): string {
    return WorkerFS.Name
  }

  public isReadOnly(): boolean {
    return this._isReadOnly
  }

  public supportsSynch(): boolean {
    return false
  }

  public supportsLinks(): boolean {
    return this._supportLinks
  }

  public supportsProps(): boolean {
    return this._supportProps
  }

  public rename(_oldPath: string, _newPath: string, _cb: BFSOneArgCallback): void {
    this._rpc('rename', arguments)
  }

  public stat(_p: string, _isLstat: boolean, _cb: BFSCallback<Stats>): void {
    this._rpc('stat', arguments)
  }

  public open(_p: string, _flag: FileFlag, _mode: number, _cb: BFSCallback<File>): void {
    this._rpc('open', arguments)
  }

  public unlink(_p: string, _cb: Function): void {
    this._rpc('unlink', arguments)
  }

  public rmdir(_p: string, _cb: Function): void {
    this._rpc('rmdir', arguments)
  }

  public mkdir(_p: string, _mode: number, _cb: Function): void {
    this._rpc('mkdir', arguments)
  }

  public readdir(_p: string, _cb: BFSCallback<string[]>): void {
    this._rpc('readdir', arguments)
  }

  public exists(_p: string, _cb: (exists: boolean) => void): void {
    this._rpc('exists', arguments)
  }

  public realpath(_p: string, _cache: Record<string, string>, _cb: BFSCallback<string>): void {
    this._rpc('realpath', arguments)
  }

  public truncate(_p: string, _len: number, _cb: Function): void {
    this._rpc('truncate', arguments)
  }

  public readFile(_fname: string, _encoding: string, _flag: FileFlag, _cb: BFSCallback<any>): void {
    this._rpc('readFile', arguments)
  }

  public writeFile(
    _fname: string,
    _data: any,
    _encoding: string,
    _flag: FileFlag,
    _mode: number,
    _cb: BFSOneArgCallback,
  ): void {
    this._rpc('writeFile', arguments)
  }

  public appendFile(
    _fname: string,
    _data: any,
    _encoding: string,
    _flag: FileFlag,
    _mode: number,
    _cb: BFSOneArgCallback,
  ): void {
    this._rpc('appendFile', arguments)
  }

  public chmod(_p: string, _isLchmod: boolean, _mode: number, _cb: Function): void {
    this._rpc('chmod', arguments)
  }

  public chown(_p: string, _isLchown: boolean, _uid: number, _gid: number, _cb: Function): void {
    this._rpc('chown', arguments)
  }

  public utimes(_p: string, _atime: Date, _mtime: Date, _cb: Function): void {
    this._rpc('utimes', arguments)
  }

  public link(_srcpath: string, _dstpath: string, _cb: Function): void {
    this._rpc('link', arguments)
  }

  public symlink(_srcpath: string, _dstpath: string, _type: string, _cb: Function): void {
    this._rpc('symlink', arguments)
  }

  public readlink(_p: string, _cb: Function): void {
    this._rpc('readlink', arguments)
  }

  public syncClose(method: string, fd: File, cb: BFSOneArgCallback): void {
    this._worker.postMessage({
      browserfsMessage: true,
      method,
      args: [(fd as unknown as WorkerFile).toRemoteArg(), this._callbackConverter.toRemoteArg(cb)],
    })
  }

  /**
   * Called once both local and remote sides are set up.
   */
  private _initialize(cb: () => void): void {
    if (!this._isInitialized) {
      const message: IAPIRequest = {
        browserfsMessage: true,
        method: 'probe',
        args: [
          this._argLocal2Remote(emptyBuffer()),
          this._callbackConverter.toRemoteArg((probeResponse: IProbeResponse) => {
            this._isInitialized = true
            this._isReadOnly = probeResponse.isReadOnly
            this._supportLinks = probeResponse.supportsLinks
            this._supportProps = probeResponse.supportsProps
            cb()
          }),
        ],
      }
      this._worker.postMessage(message)
    } else {
      cb()
    }
  }

  private _argRemote2Local(arg: any): any {
    if (!arg) {
      return arg
    }
    switch (typeof arg) {
      case 'object':
        if (typeof arg.type === 'number') {
          const specialArg = arg as ISpecialArgument
          switch (specialArg.type) {
            case SpecialArgType.API_ERROR:
              return apiErrorRemote2Local(specialArg as IAPIErrorArgument)
            case SpecialArgType.FD: {
              const fdArg = specialArg as IFileDescriptorArgument
              return new WorkerFile(
                this,
                fdArg.path,
                FileFlag.getFileFlag(fdArg.flag),
                Stats.fromBuffer(transferrableObjectToBuffer(fdArg.stat)),
                fdArg.id,
                transferrableObjectToBuffer(fdArg.data),
              )
            }
            case SpecialArgType.STATS:
              return statsRemote2Local(specialArg as IStatsArgument)
            case SpecialArgType.FILEFLAG:
              return fileFlagRemote2Local(specialArg as IFileFlagArgument)
            case SpecialArgType.BUFFER:
              return bufferRemote2Local(specialArg as IBufferArgument)
            case SpecialArgType.ERROR:
              return errorRemote2Local(specialArg as IErrorArgument)
            default:
              return arg
          }
        } else {
          return arg
        }
      default:
        return arg
    }
  }

  private _rpc(methodName: string, args: IArguments) {
    const fixedArgs = new Array(args.length)
    for (let i = 0; i < args.length; i++) {
      fixedArgs[i] = this._argLocal2Remote(args[i])
    }
    const message: IAPIRequest = {
      browserfsMessage: true,
      method: methodName,
      args: fixedArgs,
    }
    this._worker.postMessage(message)
  }

  /**
   * Converts a local argument into a remote argument. Public so WorkerFile objects can call it.
   */
  private _argLocal2Remote(arg: any): any {
    // CHANGE: here is the stuff breaks
    if (!arg) {
      return arg
    }
    switch (typeof arg) {
      case 'object':
        if (arg instanceof Stats) {
          return statsLocal2Remote(arg)
        } else if (arg instanceof ApiError) {
          return apiErrorLocal2Remote(arg)
        } else if (arg instanceof WorkerFile) {
          return arg.toRemoteArg()
        } else if (arg instanceof FileFlag) {
          return fileFlagLocal2Remote(arg)
        } else if (arg instanceof Buffer) {
          return bufferLocal2Remote(arg)
        } else if (arg instanceof Error) {
          return errorLocal2Remote(arg)
        } else {
          // CHANGE: throw an error so we know we screwed.
          throw new Error('cannot marshal FS call argument: unknown argument')
        }
      case 'function':
        return this._callbackConverter.toRemoteArg(arg)
      default:
        return arg
    }
  }
}
