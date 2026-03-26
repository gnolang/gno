import * as Comlink from 'comlink'

import type { LSPWorkerInterface, LSPWorkerStartParams } from '../common/types'
import { ServerController } from './controller'

let ctrlInstance: ServerController | null = null

/**
 * Comlink worker request handler.
 */
const handler: LSPWorkerInterface = {
  start: async (params: LSPWorkerStartParams) => {
    if (ctrlInstance) {
      return
    }

    console.log('%c[lsp] starting gnopls...', 'color: orange')
    ctrlInstance = await ServerController.create(params)
    await ctrlInstance.start()
  },
  [Comlink.releaseProxy]: async () => {
    ctrlInstance?.dispose()
    self.close()
  },
}

Comlink.expose(handler)
