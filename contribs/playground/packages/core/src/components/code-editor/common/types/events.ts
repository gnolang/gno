export enum EditorMessageSeverity {
  Debug,
  Log,
  Info,
  Warning,
  Error,
}

export enum EditorEventType {
  /**
   * Empty value. Triggered to clear previously set status bar contents.
   */
  None,

  /**
   * Resource loading or background job start event.
   */
  Progress,

  /**
   * Generic log message event.
   */
  Notification,
}

export interface EditorNotificationEvent {
  type: EditorEventType.Notification
  tag?: string
  severity: EditorMessageSeverity
  message: string
  data?: any
}

export interface EditorProgressEvent {
  type: EditorEventType.Progress
  finished?: boolean
  tag?: string
  message?: string
  percentage?: number
}

export interface EditorEmptyEvent {
  type: EditorEventType.None
}

export type EditorEvent = EditorProgressEvent | EditorNotificationEvent | EditorEmptyEvent
