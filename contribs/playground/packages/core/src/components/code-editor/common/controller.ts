import type { InputListener, InputMode } from './layout'

export type Callback<T> = (arg: T) => void
export type CursorStyle = 'line' | 'block'

export interface Size {
  height: number
  width: number
}

export interface CursorPosition {
  readonly lineNumber: number
  readonly column: number
}

export interface SelectionRange {
  /**
   * Line number on which the range starts (starts at 1).
   */
  readonly startLineNumber: number
  /**
   * Column on which the range starts in line `startLineNumber` (starts at 1).
   */
  readonly startColumn: number
  /**
   * Line number on which the range ends.
   */
  readonly endLineNumber: number
  /**
   * Column on which the range ends in line `endLineNumber`.
   */
  readonly endColumn: number
}

/**
 * EditorController interface provides offscreen control over editor instance.
 * This allows to control editor behavior outside of React (e.g. in MobX) without importing editor lib.
 *
 * Goal of interface is to abstract underlying implementation details away from business logic.
 * This is used to simplify migration from Monaco editor.
 *
 * For sake of migration convenience, most of used interfaces are derived from Monaco.
 *
 * Each editor implementation provides it's own controller.
 */
export interface EditorController {
  get doc(): string

  /**
   * Cursor position change listener.
   */
  onCursorPositionChange?: Callback<CursorPosition>

  /**
   * Gutter (line number column) click listener.
   */
  onGutterClick?: Callback<SelectionRange>

  /**
   * Input mode change listener.
   * Fired when mode changed between vim and emacs.
   *
   * Null value is passed when mode switched back to regular.
   */
  onInputModeChange?: Callback<InputListener | null>

  /**
   * Event listener to track whether editor is blocked by loading some resources (for example input method extensions).
   *
   * Clients expect to block any attached UI when editor is loading.
   */
  onLoadingStateChange?: Callback<boolean>

  /**
   * Applies document formatting.
   */
  formatDocument: () => void

  /**
   * Focuses editor instance.
   */
  focus: () => void

  /**
   * Highlights a line of text in a document.
   */
  highlightLine: (lineNumber: number) => void

  /**
   * Removes line highlighted by `highlightLine`.
   */
  removeHighlight: () => void

  /**
   * Resizes editor container. Works only in monaco and probably should be removed.
   */
  resize?: (size: Size) => void

  /**
   * Changes keybindings layout.
   * Supported only for Monaco, for CodeMirror just pass `inputMode` value into props.
   *
   * @param layout new layout type.
   */
  setInputMode?: (layout: InputMode) => Promise<void>

  /**
   * Restarts a language server client (if any is running)
   */
  restartLanguageServer: () => void

  /**
   * Detach from editor instance and free resources.
   */
  dispose: () => void
}
