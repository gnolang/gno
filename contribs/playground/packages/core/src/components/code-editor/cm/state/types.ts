import type { Diagnostic } from '@codemirror/lint'

import type { InputMode } from '../../common'
import { defaultSyntax, type Syntax } from '../extensions/syntax'

/**
 * StateData keeps per-file state information to sync applied React component props and CM extensions
 * with EditorState instance of a file.
 *
 * React component toggles CodeMirror extensions using data in StateData.
 */
export interface StateData {
  /**
   * Field to identify whether state is not initial state.
   */
  isInitialised?: boolean

  /**
   * Whether file contents are not available yet.
   */
  isLoading?: boolean

  /**
   * Document name associated with this editor state.
   */
  fileName?: string

  /**
   * Editor theme used when EditorState was created.
   * Used to track if theme extension for document state needs to be updated.
   */
  theme: 'dark' | 'light'

  /**
   * Identifies whether document view is in read-only mode.
   */
  readOnly?: boolean

  /**
   * Input method used when editor was created.
   */
  inputMode: InputMode

  /**
   * Language mode used for current document.
   *
   * Used by React component to track if syntax highlight extension should be changed.
   */
  syntax: Syntax

  /**
   * List of diagnostics for a linter plugin.
   */
  diagnostics?: Diagnostic[]
}

/**
 * Defaults for empty StateData state field.
 * Editor component should replace it right with actual values after EditorState is created.
 */
export const defaultStateData: StateData = {
  theme: 'light',
  syntax: defaultSyntax,
  inputMode: 'classic',
}
