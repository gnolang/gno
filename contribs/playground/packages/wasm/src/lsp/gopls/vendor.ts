/**
 * Custom LSP notifications.
 */
export enum CustomNotificationTypes {
  /**
   * Gno imports list changed notification.
   */
  DidImportsChanged = 'gopls/gnostudio/didImportsChanged',
}

export interface DidImportsChangedNotificationParams {
  /**
   * Workspace directory
   */
  cwd: string

  /**
   * Whether package driver detected any unresolved import.
   */
  hasErrors: boolean
}
