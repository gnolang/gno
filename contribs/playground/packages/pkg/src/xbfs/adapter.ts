import * as path from 'path'
import type { ApiError } from 'browserfs/dist/node/core/api_error'
import { BaseFileSystem, type FileSystem } from 'browserfs/dist/node/core/file_system'

const trimSuffix = (s: string, suffix: string) => (s.endsWith(suffix) ? s.slice(0, -suffix.length) : s)

/**
 * This is a reimplementation of FolderAdapter from BrowserFS but with fixed trailing bug.
 *
 * The original FolderAdapter passed invalid empty path to the wrapped file system, which caused ENOENT errors.
 * To avoid this, we trim the trailing slash from the path before passing it to the wrapped file system.
 *
 * @see: browserfs/src/backend/FolderAdapter.ts
 */
export class FolderAdapter extends BaseFileSystem implements FileSystem {
  static readonly Name = 'FolderAdapter'

  static isAvailable(): boolean {
    return true
  }

  public _wrapped: FileSystem
  public _folder: string

  /**
   * Wraps a file system, and uses the given folder as its root.
   *
   * @param folder The folder to use as the root directory.
   * @param wrapped The file system to wrap.
   */
  constructor(folder: string, wrapped: FileSystem) {
    super()
    this._folder = folder
    this._wrapped = wrapped
  }

  getName(): string {
    return this._wrapped.getName()
  }

  isReadOnly(): boolean {
    return this._wrapped.isReadOnly()
  }

  supportsProps(): boolean {
    return this._wrapped.supportsProps()
  }

  supportsSynch(): boolean {
    return this._wrapped.supportsSynch()
  }

  supportsLinks(): boolean {
    return false
  }
}

function translateError(folder: string, e: any): any {
  if (e !== null && typeof e === 'object') {
    const err = e as ApiError
    let p = err.path
    if (p) {
      p = '/' + path.relative(folder, p)

      err.message = err.message.replace(err.path!, p)
      err.path = p
    }
  }
  return e
}

function wrapCallback(folder: string, cb: any): any {
  if (typeof cb === 'function') {
    return function (err: ApiError, ...args: any[]) {
      if (arguments.length > 0) {
        err = translateError(folder, err)
      }
      cb(err, ...args)
    }
  } else {
    return cb
  }
}

// eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
function wrapFunction(name: string, wrapFirst: boolean, wrapSecond: boolean): Function {
  if (name.slice(name.length - 4) !== 'Sync') {
    // Async function. Translate error in callback.
    return function (this: FolderAdapter, ...args: any) {
      if (args.length > 0) {
        if (wrapFirst) {
          args[0] = trimSuffix(path.join(this._folder, args[0]), '/')
        }
        if (wrapSecond) {
          args[1] = trimSuffix(path.join(this._folder, args[1]), '/')
        }
        args[args.length - 1] = wrapCallback(this._folder, args[args.length - 1])
      }
      return (this._wrapped as any)[name].apply(this._wrapped, args)
    }
  }

  // Sync function. Translate error in catch.
  return function (this: FolderAdapter, ...args: any) {
    try {
      if (wrapFirst) {
        args[0] = trimSuffix(path.join(this._folder, args[0]), '/')
      }
      if (wrapSecond) {
        args[1] = path.join(this._folder, args[1])
      }
      return (this._wrapped as any)[name].apply(this._wrapped, args)
    } catch (e) {
      throw translateError(this._folder, e)
    }
  }
}

// First argument is a path.
;[
  'diskSpace',
  'stat',
  'statSync',
  'open',
  'openSync',
  'unlink',
  'unlinkSync',
  'rmdir',
  'rmdirSync',
  'mkdir',
  'mkdirSync',
  'readdir',
  'readdirSync',
  'exists',
  'existsSync',
  'realpath',
  'realpathSync',
  'truncate',
  'truncateSync',
  'readFile',
  'readFileSync',
  'writeFile',
  'writeFileSync',
  'appendFile',
  'appendFileSync',
  'chmod',
  'chmodSync',
  'chown',
  'chownSync',
  'utimes',
  'utimesSync',
  'readlink',
  'readlinkSync',
].forEach((name: string) => {
  ;(FolderAdapter.prototype as any)[name] = wrapFunction(name, true, false)
})

// First and second arguments are paths.
;['rename', 'renameSync', 'link', 'linkSync', 'symlink', 'symlinkSync'].forEach((name: string) => {
  ;(FolderAdapter.prototype as any)[name] = wrapFunction(name, true, true)
})
