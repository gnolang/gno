/**
 * SyscallError is Error type mapped to syscall.Errno type by syscall/js.
 *
 * Unlike BFS's ApiError, supports custom error codes which aren't present in ApiError.
 */
export class SyscallError extends Error {
  constructor(
    public code: string,
    message: string,
    public path?: string,
  ) {
    super(message)
  }

  static ENOSYS() {
    return new SyscallError('ENOSYS', 'not implemented')
  }

  static ENOENT(path?: string) {
    return new SyscallError('ENOENT', 'no such file or directory', path)
  }

  static EBADF() {
    return new SyscallError('EBADF', 'bad file descriptor')
  }

  static EISDIR() {
    return new SyscallError('EISDIR', 'file is a directory')
  }

  /**
   * EIO is input/output error designed for internal, unexpected errors.
   */
  static EIO(message?: string) {
    return new SyscallError('EIO', message ?? 'internal error')
  }

  /**
   * Ensures that error is a correct file system error, otherwize returns a proper error with code.
   *
   * Context:
   * Go file system can only handle JS errors that have a `code` property with a correct value.
   * If error doesn't have a code, Go program just panics with `<object>` message.
   *
   * Usually such kind of errors signal about critical internal error and should be handled before.
   * For this case, this method also logs such errors in console with stack trace but still keep Go programs running.
   *
   * @param err
   */
  static normalizeError(err: any | null | undefined) {
    if (!err) {
      // Go expects empty errors to be 'null', not undefined.
      // See: $GOROOT/src/syscall/fs_js.go:519
      return null
    }

    if ('code' in err) {
      return err
    }

    // Log obscure error as Go won't log a whole error context.
    console.error('wasmfs: unexpected error', err)
    return SyscallError.EIO(err.message)
  }
}
