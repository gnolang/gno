import { attachWorkerFSListener, KeyValueFileSystem, type FilesMap } from '@gnostudio/pkg'

import * as BrowserFS from 'browserfs'
import * as Comlink from 'comlink'

import type { LSPWorkerInterface } from './types'

export interface LSPStartParams {
  /**
   * Enable request debug logging.
   */
  debug?: boolean

  /**
   * Whether returned JSON-RPC response should be a raw JSON string or object.
   */
  responseEncoding: 'json' | 'object'

  workspace: {
    path: string
    files: FilesMap
  }
}

/**
 * Set of ports for FS and LSP requests communication.
 *
 * First port is used by client, second is transfered to a worker.
 */
interface MessageChannels {
  lsp: MessageChannel
  fs: MessageChannel
}

/**
 * Wrapper around comlink object that abstracts file system creation and transferrable arguments passing.
 */
export class LSPWorkerClient {
  private readonly instance: LSPWorkerInterface
  private readonly channels: MessageChannels
  private disposed = false

  /**
   * Returns message port to send/receive requests for LSP client.
   */
  get lspPort() {
    return this.channels.lsp.port1
  }

  constructor(private readonly worker: Worker) {
    this.instance = Comlink.wrap<LSPWorkerInterface>(worker)
    this.channels = {
      lsp: new MessageChannel(),
      fs: new MessageChannel(),
    }
  }

  /**
   * Starts LSP server.
   */
  start({ debug, workspace, responseEncoding }: LSPStartParams) {
    const { port1: fsListenPort, port2: fsWorkerPort } = this.channels.fs
    const { port1: lspClientPort, port2: lspWorkerPort } = this.channels.lsp

    // Subscription from 'addEventListener' won't work unless start is called.
    // See: https://issues.chromium.org/issues/40429353
    fsListenPort.start()
    lspClientPort.start()
    const workspaceFs = new KeyValueFileSystem(workspace.files, {
      stripRootSlash: true,
    })

    // WorkerFS host is binded to a global filesystem instance.
    // Only one global BFS instance should exist at the same time.
    const bfs = BrowserFS.initialize(workspaceFs)
    attachWorkerFSListener(bfs, fsListenPort)

    const params = {
      debug,
      responseEncoding,
      workspacePath: workspace.path,
      ports: {
        lsp: lspWorkerPort,
        fs: fsWorkerPort,
      },
    }

    return this.instance.start(Comlink.transfer(params, [lspWorkerPort, fsWorkerPort]))
  }

  private closePorts() {
    this.channels.fs.port1.close()
    this.channels.lsp.port1.close()
  }

  /**
   * Terminates worker and releases all allocated resources.
   */
  dispose() {
    if (this.disposed) {
      return
    }

    this.instance[Comlink.releaseProxy]()
    this.closePorts()
    this.worker.terminate()
    this.disposed = true
  }
}
