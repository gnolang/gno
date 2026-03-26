import type {
  CompletionItem,
  CompletionList,
  CompletionParams,
  DidChangeTextDocumentParams,
  DidOpenTextDocumentParams,
  DocumentFormattingParams,
  Hover,
  HoverParams,
  InitializedParams,
  InitializeParams,
  InitializeResult,
  LogMessageParams,
  PublishDiagnosticsParams,
  ShowMessageParams,
  ShowMessageRequestParams,
  TextEdit,
  WorkDoneProgressBegin,
  WorkDoneProgressCancelParams,
  WorkDoneProgressEnd,
  WorkDoneProgressParams,
  WorkDoneProgressReport,
} from 'vscode-languageserver-protocol'

export { MessageType, type ShowMessageParams, type LogMessageParams } from 'vscode-languageserver-protocol'
export type ProgressParams = WorkDoneProgressBegin | WorkDoneProgressReport | WorkDoneProgressEnd

// Client to server then server to client
export interface LSPRequestMap {
  initialize: [InitializeParams, InitializeResult]
  'textDocument/hover': [HoverParams, Hover]
  'textDocument/completion': [CompletionParams, CompletionItem[] | CompletionList | null]
  'textDocument/didChange': [DidChangeTextDocumentParams, PublishDiagnosticsParams]
  'textDocument/formatting': [DocumentFormattingParams, TextEdit[]]
  'completionItem/resolve': [CompletionItem, CompletionItem]
}

// Client to server
export interface LSPNotifyMap {
  initialized: InitializedParams
  'textDocument/didChange': DidChangeTextDocumentParams
  'textDocument/didOpen': DidOpenTextDocumentParams
  'textDocument/formatting': DocumentFormattingParams
}

// Server to client
export interface LSPEventMap {
  'textDocument/publishDiagnostics': PublishDiagnosticsParams
  'window/showMessage': ShowMessageParams
  'window/showMessageRequest': ShowMessageRequestParams
  'window/logMessage': LogMessageParams
  'window/workDoneProgress/create': WorkDoneProgressParams
  'window/workDoneProgress/cancel': WorkDoneProgressCancelParams
  '$/progress': ProgressParams
}

export type Notification = {
  [key in keyof LSPEventMap]: {
    jsonrpc: '2.0'
    id?: null | undefined
    method: key
    params: LSPEventMap[key]
  }
}[keyof LSPEventMap]

/**
 * LSP client event subscriber.
 *
 * LSP plugin supposed to implement this interface.
 */
export interface ClientSubscriber {
  handleNotification: (notification: Notification) => void
  initialize: () => void
}

/**
 * Handles notifications and requests to application window (shell).
 *
 * Those notifications are not related to a document and has to be handled/displayed in an app shell.
 */
export interface WindowNotificationHandler {
  /**
   * Handle incoming log messages.
   */
  showMessage?: (message: ShowMessageParams | LogMessageParams) => void

  /**
   * Handle background job progress notifications.
   */
  onProgress?: (progress: ProgressParams) => void

  /**
   * Error handler for language server start error.
   */
  onInitError?: (err: Error) => void

  /**
   * Called when LSP server starts.
   */
  onInitBegin?: () => void
}
