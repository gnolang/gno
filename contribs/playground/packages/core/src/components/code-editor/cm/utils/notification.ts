import {
  MessageType,
  type LogMessageParams,
  type ProgressParams,
  type ShowMessageParams,
} from '@gnostudio/codemirror-lsp'

import {
  EditorEventType,
  EditorMessageSeverity,
  type EditorEmptyEvent,
  type EditorNotificationEvent,
  type EditorProgressEvent,
} from '../../common/types'

const severityMapping: Record<MessageType, EditorMessageSeverity> = {
  [MessageType.Debug]: EditorMessageSeverity.Debug,
  [MessageType.Log]: EditorMessageSeverity.Log,
  [MessageType.Info]: EditorMessageSeverity.Info,
  [MessageType.Warning]: EditorMessageSeverity.Warning,
  [MessageType.Error]: EditorMessageSeverity.Error,
}

export const mapLspMessage = (message: ShowMessageParams | LogMessageParams): EditorNotificationEvent => ({
  type: EditorEventType.Notification,
  tag: 'LSP',
  severity: severityMapping[message.type],
  message: message.message,
})

export const mapLspProgress = (message: ProgressParams): EditorProgressEvent => ({
  type: EditorEventType.Progress,
  tag: 'LSP',
  finished: message.kind === 'end',
  message: 'title' in message ? message.title : message.message,
  percentage: 'percentage' in message ? message.percentage : undefined,
})

export const newLspStartErrorMessage = (): EditorNotificationEvent => ({
  type: EditorEventType.Notification,
  severity: EditorMessageSeverity.Error,
  tag: 'LSP',
  message: 'Failed to start language server',
})

export const newLspStartEvent = (): EditorProgressEvent => ({
  type: EditorEventType.Progress,
  tag: 'LSP',
  message: 'Starting language server',
})

export const newTerminateEvent = (): EditorEmptyEvent => ({
  type: EditorEventType.None,
})

export const newFileLoadErrorEvent = (
  workspaceId: string,
  fileName: string,
  error: Error,
): EditorNotificationEvent => ({
  type: EditorEventType.Notification,
  message: String(error),
  severity: EditorMessageSeverity.Error,
  data: {
    workspaceId,
    fileName,
    error,
  },
})
