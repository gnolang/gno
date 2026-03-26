import { LSPWorkerClient } from './common/client'

export * from './common/client'
export * from './common/types'
export * from './gopls/vendor'

/**
 * Starts a gopls LSP server worker and returns a worker client.
 */
export const newGoplsWorker = () => {
  const worker = new Worker(new URL('./gopls/worker', import.meta.url), { type: 'module' })
  return new LSPWorkerClient(worker)
}

/**
 * Starts a gnopls LSP server worker and returns a worker client.
 */
export const newGnoplsWorker = () => {
  const worker = new Worker(new URL('./gnopls/worker', import.meta.url), { type: 'module' })
  return new LSPWorkerClient(worker)
}
