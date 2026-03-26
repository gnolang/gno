import { Client, RequestManager } from '@open-rpc/client-js'
import type { Transport } from '@open-rpc/client-js/build/transports/Transport'
import type {
  CompletionItem,
  CompletionParams,
  DidChangeTextDocumentParams,
  DidOpenTextDocumentParams,
  DocumentFormattingParams,
  HoverParams,
  LSPAny,
  ServerCapabilities,
  WorkspaceFolder,
} from 'vscode-languageserver-protocol'

import { newInitializeRequest } from './request'
import type { ClientSubscriber, LSPNotifyMap, LSPRequestMap, Notification, WindowNotificationHandler } from './types'

const requestTimeout = 10000
const initTimeout = requestTimeout * 3

export interface LanguageServerClientOptions {
  /**
   * Workspace directory URI
   */
  rootUri: string

  /**
   * List of workspace folders, alternative to rootUri.
   */
  workspaceFolders?: WorkspaceFolder[]

  /**
   * Automatically shut down a client after all plugin instances have been detached.
   */
  autoClose?: boolean

  /**
   * LSP communication transport.
   */
  transport: Transport

  /**
   * LSP server messages handler.
   */
  windowNotificationHandler?: WindowNotificationHandler

  /**
   * Additional hook to be called before client stablishes LSP connection.
   *
   * Used to implement additional bootstrap logic, such as installing third party dependencies.
   *
   * @param rootUri Workspace URI.
   */
  bootstrapHook?: (rootUri: string) => Promise<void>

  /**
   * Custom LSP notifications handlers for non-standard server notifications.
   */
  customNotifications?: Record<string, any>
}

const logRequestError = (req: string, err: any) => {
  console.error(`lsp: "${req}" failed, server returned error:`, err)
}

export class LanguageServerClient {
  private readonly plugins: ClientSubscriber[] = []
  private requestManager!: RequestManager
  private client!: Client

  public ready = false
  public capabilities?: ServerCapabilities<LSPAny>

  public initializePromise!: Promise<void>

  get rootUri() {
    return this.config.rootUri
  }

  constructor(private readonly config: LanguageServerClientOptions) {
    this.reconfigure()
  }

  private reconfigure() {
    this.requestManager = new RequestManager([this.config.transport])
    this.client = new Client(this.requestManager)
    this.client.onNotification((data) => {
      this.handleNotification(data as any)
    })

    this.initializePromise = this.initialize()
  }

  private async initialize() {
    const { windowNotificationHandler, workspaceFolders, rootUri, bootstrapHook } = this.config
    try {
      await bootstrapHook?.(rootUri)

      windowNotificationHandler?.onInitBegin?.()
      await this.requestManager.connectPromise
      const { capabilities } = await this.request(
        'initialize',
        newInitializeRequest({
          rootUri,
          workspaceFolders,
        }),
        initTimeout,
      )

      this.capabilities = capabilities
    } catch (err) {
      windowNotificationHandler?.onInitError?.(err as Error)
      throw err
    }

    await this.notify('initialized', {})
    this.ready = true
  }

  /**
   * Restarts LSP session.
   *
   * All connected client sessions are restored.
   */
  reconnect() {
    this.client.close()
    this.ready = false
    this.reconfigure()
  }

  close() {
    this.client.close()
  }

  textDocumentDidOpen(params: DidOpenTextDocumentParams) {
    return this.notify('textDocument/didOpen', params)
  }

  textDocumentDidChange(params: DidChangeTextDocumentParams) {
    return this.notify('textDocument/didChange', params)
  }

  async textDocumentFormatting(params: DocumentFormattingParams) {
    return await this.request('textDocument/formatting', params, requestTimeout)
  }

  async textDocumentLinting(params: DidChangeTextDocumentParams) {
    return await this.request('textDocument/didChange', params, requestTimeout)
  }

  async textDocumentHover(params: HoverParams) {
    try {
      return await this.request('textDocument/hover', params, requestTimeout)
    } catch (err) {
      logRequestError('textDocument/hover', err)
      return null
    }
  }

  async textDocumentCompletion(params: CompletionParams) {
    try {
      return await this.request('textDocument/completion', params, requestTimeout)
    } catch (err) {
      logRequestError('textDocument/completion', err)
      return null
    }
  }

  async completionItemResolve(item: CompletionItem) {
    try {
      return await this.request('completionItem/resolve', item, requestTimeout)
    } catch (err) {
      logRequestError('completionItem/resolve', err)
      return item
    }
  }

  attachPlugin(plugin: ClientSubscriber) {
    this.plugins.push(plugin)
  }

  detachPlugin(plugin: ClientSubscriber) {
    const i = this.plugins.indexOf(plugin)
    if (i === -1) return
    this.plugins.splice(i, 1)

    if (!this.plugins.length && this.config.autoClose) {
      this.close()
    }
  }

  private request<K extends keyof LSPRequestMap>(
    method: K,
    params: LSPRequestMap[K][0],
    timeout: number,
  ): Promise<LSPRequestMap[K][1]> {
    return this.client.request({ method, params }, timeout)
  }

  private notify<K extends keyof LSPNotifyMap>(method: K, params: LSPNotifyMap[K]): Promise<LSPNotifyMap[K]> {
    return this.client.notify({ method, params })
  }

  private handleNotification(notification: Notification) {
    switch (notification.method) {
      case 'window/logMessage':
      case 'window/showMessage':
        this.config.windowNotificationHandler?.showMessage?.(notification.params)
        return
      case '$/progress':
        this.config.windowNotificationHandler?.onProgress?.(notification.params)
        return
      default:
        break
    }

    const handler = this.config.customNotifications?.[notification.method]
    if (handler) {
      handler(notification.params)
      return
    }

    for (const plugin of this.plugins) {
      plugin.handleNotification(notification)
    }
  }
}
