/**
 * BrowserUnsupportedError indicates that browser doesn't support all required WASM features.
 *
 * @see checkWasmExtendedSupport
 */
export class BrowserUnsupportedError extends Error {
  constructor(message: string) {
    super(`Unsupported environment: ${message}`)
    this.name = 'BrowserUnsupportedError'
  }
}

/**
 * Checks whether environment support all features required by Gno WebAssembly workers.
 * Throws BrowserUnsupportedError on failure.
 *
 * @see BrowserUnsupportedError
 */
export const checkWebAssemblySupport = () => {
  if (!('ReadableStreamDefaultController' in globalThis)) {
    throw new BrowserUnsupportedError('ReadableStreamDefaultController is missing')
  }
}

export const isBrowserUnsupportedError = (err: Error) => {
  return err.name === 'BrowserUnsupportedError'
}
