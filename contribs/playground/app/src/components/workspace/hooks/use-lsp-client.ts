import { useEffect, useMemo, useRef } from 'react'

import { MessagePortTransport, type LanguageServerClientOptions } from '@gnostudio/codemirror-lsp'
import { newGnoplsWorker, newGoplsWorker, type LSPWorkerClient } from '@gnostudio/wasm'

import { useStore } from '@/contexts'

const WORKSPACE_ROOT = '/work'

export enum LSPServerType {
  Gopls,
  Gnopls,
}

export const buildWorkerClient = (serverName: LSPServerType) => {
  switch (serverName) {
    case LSPServerType.Gnopls:
      return newGnoplsWorker()
    case LSPServerType.Gopls:
      return newGoplsWorker()
    default:
      throw new Error(`unknown LSP server: ${LSPServerType[serverName]}`)
  }
}

/**
 * Hook provides LSP client configuration for CMCodeEditor.
 * LSP workspace is populated from MST store.
 */
export const useLspClient = (serverType: LSPServerType) => {
  const store = useStore()
  const clientRef = useRef<LSPWorkerClient | null>(null)

  // LSP server has to be restarted every time project changes.
  const projectId: string = store.workbench.workspaceId

  // TODO: add lsp boot timeout?
  const lspConfig: LanguageServerClientOptions = useMemo(() => {
    clientRef.current?.dispose()

    const workDir = `${WORKSPACE_ROOT}/${projectId}`
    const rootUri = `file://${workDir}`

    // Start gnopls only when LSP client is started.
    const portProvider = async () => {
      const client = buildWorkerClient(serverType)
      clientRef.current = client

      await client.start({
        debug: !!import.meta.env.VITE_LSP_DEBUG,
        responseEncoding: 'json',
        workspace: {
          path: workDir,
          files: store.workbench.files,
        },
      })

      return client.lspPort
    }

    return {
      rootUri,
      bootstrapHook: async () => {
        // TODO: run "gno mod download"
      },
      transport: new MessagePortTransport(portProvider, {
        onClose: () => clientRef.current?.dispose(),
      }),
      // TODO: Uncomment when `bootstrapHook` will be implemented!
      // customNotifications: {
      //   [CustomNotificationTypes.DidImportsChanged]: ({ hasErrors }: DidImportsChangedNotificationParams) => {
      //     if (!hasErrors) {
      //       return
      //     }

      //     // "gno mod download" will be triggered by bootstrapHook.
      //     store.editor.editorCtrl?.restartLanguageServer()
      //   },
      // },
    }
  }, [store, serverType, projectId])

  useEffect(() => {
    // Shut down gnopls if not killed after editor unmount.
    return () => clientRef.current?.dispose()
  }, [])

  return lspConfig
}
