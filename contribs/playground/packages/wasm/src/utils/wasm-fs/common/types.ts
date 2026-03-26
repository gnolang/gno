import type { ApiError } from 'browserfs/dist/node/core/api_error'
import type BFSStats from 'browserfs/dist/node/core/node_fs_stats'

import { SyscallError } from './errors'

// Helper types, similar to BFS types except that it accept custom error types.
export type ErrorType = SyscallError | ApiError | Error | null
export type ErrCallback = (e?: ErrorType) => any
export type ResultCallback<T> = (e?: ErrorType, rv?: T) => any
export type ThreeArgCallback<T, U> = (e?: ErrorType, arg1?: T, arg2?: U) => any

/**
 * Go-specific file stats format.
 */
export interface GoFileStats extends BFSStats {
  atimeMs: number
  mtimeMs: number
  ctimeMs: number
}

/**
 * Wraps error callback with Go error normalizer.
 *
 * @see SyscallError.normalizeError
 */
export const wrapErrCallback = (cb: ErrCallback): ErrCallback => {
  return (err) => cb(SyscallError.normalizeError(err))
}

/**
 * Wraps result callback with Go error normalizer.
 *
 * @see SyscallError.normalizeError
 */
export const wrapResultCallback = <T>(cb: ResultCallback<T>): ResultCallback<T> => {
  return (err?: ErrorType, result?: T) => cb(SyscallError.normalizeError(err), result)
}

/**
 * Wraps 3-arg callback with Go error normalizer.
 *
 * @see SyscallError.normalizeError
 */
export const wrapThreeArgCallback = <T, U>(cb: ThreeArgCallback<T, U>): ThreeArgCallback<T, U> => {
  return (err?: ErrorType, arg1?: T, arg2?: U) => cb(SyscallError.normalizeError(err), arg1, arg2)
}
