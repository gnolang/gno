import type { releaseProxy } from 'comlink'

/**
 * Common LSP server worker startup params.
 */
export interface LSPWorkerStartParams {
  /**
   * Enable debug log and tracing.
   */
  debug?: boolean

  /**
   * Path where project directory is located.
   */
  workspacePath: string

  /**
   * Whether returned JSON-RPC response should be a raw JSON string or object.
   */
  responseEncoding: 'json' | 'object'

  ports: {
    /**
     * Message port for lsp requests.
     */
    lsp: MessagePort

    /**
     * Port for remote file system calls.
     */
    fs: MessagePort
  }
}

/**
 * Common LSP server worker communication interface exposed by Comlink.
 */
export interface LSPWorkerInterface {
  /**
   * Starts a gopls process.
   */
  start: (params: LSPWorkerStartParams) => Promise<void>

  /**
   * Destructor method called when Comlink proxy shuts down.
   */
  [releaseProxy]: () => void
}
