import { ApiError, ErrorCode } from 'browserfs/dist/node/core/api_error'

export const decoder = new TextDecoder()
export const encoder = new TextEncoder()

export const truncateString = (str: string, len: number) =>
  str.length >= len ? str.slice(0, len) : str + '\0'.repeat(len - str.length)

export const newApiError = (code: ErrorCode): ApiError => {
  // HACK: for some reason, `ApiError` prototype is lost and replaced with regular `Error`.
  const err = new ApiError(code)
  Object.setPrototypeOf(err, ApiError.prototype)
  return err
}

export const eReadOnly = () => newApiError(ErrorCode.EROFS)
export const enoent = () => newApiError(ErrorCode.ENOENT)
