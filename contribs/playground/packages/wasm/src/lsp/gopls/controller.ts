import type { RequestMessage } from 'vscode-languageserver-protocol'

import { loadString, type Go } from '../../go/wasm-exec'
import { WasmService } from '../../service'
import { BaseServerController } from '../common/controller'
import type { LSPWorkerStartParams } from '../common/types'
import { env, GOPATH, goplsUrl, GOROOT, ignoredRequests } from './config'
import { createRootFs } from './fs'
import { CustomNotificationTypes, type DidImportsChangedNotificationParams } from './vendor'

/**
 * Debounce timer interval in milliseconds for imports change notification.
 */
const PKG_DRIVER_HOOK_DEBOUNCE_INTERVAL = 1000

/**
 * Server controller implementation for gopls.
 */
export class ServerController extends BaseServerController {
  private isPackageDriverInitialized = false
  private prevTimeoutId: ReturnType<typeof setTimeout> | null = null

  protected getEnvironmentVariables(): Record<string, string> {
    return env
  }

  protected isRequestIgnored(req: RequestMessage): boolean {
    return ignoredRequests.has(req.method)
  }

  static async create(params: LSPWorkerStartParams) {
    if (!goplsUrl) {
      throw new Error('gopls URL is not defined')
    }

    // Gnopls might misbehave if gno.mod file is missing.
    const fs = await createRootFs(params.ports.fs, {
      GOPATH,
      GOROOT,
      workspaceDir: params.workspacePath,
    })
    const goplsService = new WasmService(goplsUrl)
    return new ServerController(fs, goplsService, params)
  }

  protected setupImports(go: Go): void {
    super.setupImports(go)

    go.importObject.lsp.notifyPackageDriverCalled = (cwdStrPtr: number, errFlag: number) => {
      const cwd = loadString(go.mem, cwdStrPtr)

      if (this.params.debug) {
        console.log('%c[lsp] %cPackage driver hook called', 'color:orange', 'color:gray', cwd)
      }

      // WASM doesn't support booleans.
      const hasErrors = errFlag > 0
      this.handlePackageDriverHook(cwd, hasErrors)
    }
  }

  protected handlePackageDriverHook(cwd: string, hasErrors: boolean): void {
    if (!this.isPackageDriverInitialized) {
      // First package driver call is done right after server boot.
      // Skip call 'gno mod download' has been already called before starting LSP server.
      this.isPackageDriverInitialized = true
      return
    }

    // Debounce call as event fired during every keystroke inside imports.
    if (this.prevTimeoutId) {
      clearTimeout(this.prevTimeoutId)
    }

    this.prevTimeoutId = setTimeout(() => {
      this.postMessage({
        jsonrpc: '2.0',
        method: CustomNotificationTypes.DidImportsChanged,
        params: {
          cwd,
          hasErrors,
        } satisfies DidImportsChangedNotificationParams,
      })
    }, PKG_DRIVER_HOOK_DEBOUNCE_INTERVAL)
  }
}
