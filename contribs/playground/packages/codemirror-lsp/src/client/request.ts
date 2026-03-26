import type { InitializeParams, WorkspaceFolder } from 'vscode-languageserver-protocol'

interface InitializeContext {
  rootUri: string | null
  workspaceFolders?: WorkspaceFolder[]
}

export const newInitializeRequest = ({ rootUri, workspaceFolders }: InitializeContext): InitializeParams => ({
  clientInfo: {
    name: 'gnostudio.codemirror-lsp',
    version: '1.0.0',
  },
  capabilities: {
    textDocument: {
      hover: {
        dynamicRegistration: true,
        contentFormat: ['markdown', 'plaintext'],
      },
      moniker: {},
      synchronization: {
        dynamicRegistration: true,
        willSave: false,
        didSave: false,
        willSaveWaitUntil: false,
      },
      completion: {
        dynamicRegistration: true,
        completionItem: {
          snippetSupport: false,
          commitCharactersSupport: true,
          documentationFormat: ['markdown', 'plaintext'],
          deprecatedSupport: false,
          preselectSupport: false,
        },
        contextSupport: false,
      },
      signatureHelp: {
        dynamicRegistration: true,
        signatureInformation: {
          documentationFormat: ['markdown', 'plaintext'],
        },
      },
      declaration: {
        dynamicRegistration: true,
        linkSupport: true,
      },
      definition: {
        dynamicRegistration: true,
        linkSupport: true,
      },
      typeDefinition: {
        dynamicRegistration: true,
        linkSupport: true,
      },
      implementation: {
        dynamicRegistration: true,
        linkSupport: true,
      },
    },
    workspace: {
      didChangeConfiguration: {
        dynamicRegistration: true,
      },
    },
  },
  initializationOptions: null,
  processId: null,
  rootUri,
  workspaceFolders,
})
