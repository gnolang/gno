import type { NotificationMessage, RequestMessage, ResponseMessage } from 'vscode-languageserver-protocol'

import { Go, loadString } from '../../go/wasm-exec'
import { type WasmService } from '../../service'
import type { WasmFS } from '../../utils/wasm-fs'
import type { LSPWorkerStartParams } from './types'

type RequestHandlerFunc = (message: string) => void

/**
 * Base class to implement LSP server controller.
 *
 * Responsible for server startup, configuration and handling LSP requests.
 */
export abstract class BaseServerController {
  private isRunning = false
  private requestHandler: RequestHandlerFunc | null = null
  private readonly lspPort: MessagePort

  protected constructor(
    private readonly fs: WasmFS,
    private readonly wasmSvc: WasmService,
    protected readonly params: LSPWorkerStartParams,
  ) {
    this.lspPort = params.ports.lsp
  }

  /**
   * Returns environment variables for LSP process.
   */
  protected abstract getEnvironmentVariables(): Record<string, string>

  protected isRequestIgnored(_req: RequestMessage): boolean {
    return false
  }

  async start() {
    if (this.isRunning) {
      console.warn('lsp: another instance is already running, skip init.')
      return
    }

    await this.wasmSvc.fetchWasmModule()
    const go = new Go()
    go.vars.fs = this.fs

    // See: https://github.com/gnostudio/studio/issues/574
    go.vars.fetch = (...args: Parameters<typeof globalThis.fetch>) => fetch.apply(globalThis, args)

    // Set workspace to be working directory.
    const { process } = go.vars
    go.vars.process = {
      ...process,
      cwd: () => this.params.workspacePath,
    }

    this.setupImports(go)
    const instance = await this.wasmSvc.load(go.importObject)

    this.lspPort.onmessage = (e: MessageEvent<RequestMessage>) => {
      this.handleRequest(e.data).catch(console.error)
    }

    this.isRunning = true
    go.run(instance)
      .then(() => console.log('%c[lsp] worker initialized', 'color: orange'))
      .catch((err) => {
        console.error('lsp worker returned an error:', err)
      })
      .finally(() => {
        this.isRunning = false
      })
  }

  dispose() {
    // We might want to reuse message channel in the future.
    // this.port.close()
  }

  private async handleRequest(req: RequestMessage) {
    if (!this.isRunning) {
      console.log('lsp: client not running, starting...')
      await this.start()
    }

    if (!this.requestHandler) {
      console.warn('lsp: request handler not initialized yet!')
    }

    if (this.params.debug) {
      const { jsonrpc: _, ...logBody } = req
      console.log('%c[lsp] %cReceive', 'color: orange', 'color: green', logBody)
    }

    if (this.isRequestIgnored(req)) {
      const rsp = { jsonrpc: req.jsonrpc, id: req.id, result: null }
      const respondAsString = this.params.responseEncoding === 'json'
      this.lspPort.postMessage(respondAsString ? JSON.stringify(rsp) : rsp)

      if (this.params.debug) {
        console.log('%c[lsp] %cReply', 'color: orange', 'color: dodgerblue', rsp)
      }
      return
    }

    const content = JSON.stringify(req)
    const body = `Content-Length: ${content.length}\r\n\r\n${content}`

    this.requestHandler?.(body)
  }

  private sendRawResponse(rawMsg: string) {
    const respondAsString = this.params.responseEncoding === 'json'

    if (this.params.debug) {
      const { jsonrpc: _, ...logBody } = JSON.parse(rawMsg)
      console.log('%c[lsp] %cReply', 'color: orange', 'color: dodgerblue', logBody)
    }

    if (respondAsString) {
      this.lspPort.postMessage(rawMsg)
      return
    }

    const msg: ResponseMessage & RequestMessage = JSON.parse(rawMsg)
    this.lspPort.postMessage(msg)
  }

  /**
   * Send a message back to LSP client on behalf of a server.
   */
  protected postMessage(msg: NotificationMessage | RequestMessage) {
    const respondAsString = this.params.responseEncoding === 'json'
    if (this.params.debug) {
      console.log('%c[lsp] %cReply', 'color: orange', 'color: dodgerblue', msg)
    }

    this.lspPort.postMessage(respondAsString ? JSON.stringify(msg) : msg)
  }

  /**
   * Configures import object and environment variables.
   *
   * Overriden implementations should call parent implementation before custom logic.
   */
  protected setupImports(go: Go) {
    const envVars = this.getEnvironmentVariables()
    go.env = {
      ...envVars,
      LSP_DEBUG: this.params.debug ? '1' : '0',
    }

    // Setup LSP request read & write hooks.
    // See: https://github.com/gnolang/gnopls/blob/master/internal/js/imports_js.go
    go.importObject.lsp = {
      writeMessage: (slicePtr: number) => {
        let rawMsg = ''
        try {
          rawMsg = loadString(go.mem, slicePtr)

          // Skip header.
          if (rawMsg.startsWith('Content-Length:')) {
            return
          }

          this.sendRawResponse(rawMsg)
        } catch (err) {
          console.error(`lsp.writeMessage: failed to handle LSP message - ${String(err)}`, {
            msg: rawMsg,
            sp: slicePtr,
          })
        }
      },
      debugLog: (sp: number) => {
        if (!this.params.debug) {
          return
        }

        const str = loadString(go.mem, sp)
        console.log('%c[lsp] %cDebug', 'color:orange', 'color:gray', str)
      },
      closeWriter: (sp: number) => {
        console.log('lsp.lspCloseWriter', { sp })
      },
      registerCallback: (callbackId: number) => {
        if (this.params.debug) {
          console.log('lsp.registerCallback: registered server callback', callbackId)
        }

        this.requestHandler = go._makeFuncWrapper(callbackId)
      },
    }
  }
}
