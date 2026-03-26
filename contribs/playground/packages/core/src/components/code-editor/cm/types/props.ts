import type { LanguageServerClientOptions } from '@gnostudio/codemirror-lsp'

import type { ContextMenuAction, EditorController, EditorEvent, InputMode } from '../../common'
import { type onRunCallback } from '../extensions/run-gutter'
import type { EditorStateManager } from '../manager'

export interface Marker {
  line: number
  column: number
  message: string
  severity?: 'info' | 'warning' | 'error'
}

export interface FormatResult {
  text?: string
  markers?: Marker[]
}

export interface File {
  /**
   * Absolute file path.
   *
   * Used to associate per-file editor state as plugins or cursor position.
   */
  path: string

  /**
   * Function or value to retreive initial value.
   *
   * Used once to populate initial editor contents.
   */
  content: string | ((fileName: string) => Promise<string>)
}

export type ThemeName = 'light' | 'dark'

export interface EditorProps {
  /**
   * Editor theme.
   */
  theme?: ThemeName

  /**
   * Whether content is read-only and not editable.
   */
  readonly?: boolean

  /**
   * External per-file view state manager.
   *
   * Consumers can provide external instance to remove unused instances.
   */
  manager?: EditorStateManager

  /**
   * Workspace identifier to flush stored per-file editor state cache.
   *
   * Changed value tells that current workspace has been changed and contents should be invalidated.
   * Update this value to reset editor state cache when file contents need to be updated.
   */
  workspaceId?: string | null

  /**
   * Active file name and contents.
   *
   * File content is used only for a first time when editor state is created for a specific file name.
   * On next re-render, content won't be used unless `workspaceId` is updated.
   *
   * When file content is async function, content load error events can be observed with `onEvent` handler.
   */
  value?: File

  /**
   * Controls whether editor should use regular or vim/emacs input mode.
   *
   * Regular is default mode.
   */
  inputMode?: InputMode

  /**
   * Whether to show line numbers.
   */
  showLineNumbers?: boolean

  /**
   * Document format function, triggered either by a hotkey or externally via EditorController.
   */
  formatter?: (doc: string) => Promise<FormatResult>

  /**
   * LSP client configuration.
   *
   * When specified, enables integration with language server.
   */
  lsp?: Pick<LanguageServerClientOptions, 'rootUri' | 'transport' | 'bootstrapHook' | 'customNotifications'>

  /**
   * Value change handler
   */
  onChange?: (content: string, path: string) => void

  /**
   * Manual save hotkey handler.
   *
   * Function can return a promise with new contents to update editor state.
   * This is can be used to implement on before save operations such as formatting.
   */
  onSave?: ((content: string) => Promise<string>) | ((content: string) => void)

  /**
   * Component mount hook
   * @param ctrl Editor remote controller
   * @see EditorController
   */
  onMount?: (ctrl: EditorController) => void

  /**
   * Event handlers for context menu actions.
   * Kept for compatibility with monaco editor interface.
   */
  onContextMenuAction?: (action: ContextMenuAction) => void

  /**
   * Action to be performed when a function is executed on the gutter marker.
   */
  onRunFunction?: onRunCallback
  /**
   * Miscellaneous editor events handler.
   *
   * Used to react on background editor events, such as loading LSP client.
   */
  onEvent?: (event: EditorEvent) => void
}
