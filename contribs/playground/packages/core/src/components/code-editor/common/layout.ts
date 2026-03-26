/**
 * Input layout type.
 */
export type InputMode = 'classic' | 'vim' | 'emacs'

type CallbackSlot<T> = (callback: T) => void

interface BaseInputListener {
  /**
   * Attaches input listener to editor to start intercepting input.
   *
   * Optional and present only on implementations that require explicit attachment.
   */
  attach?: () => void
}

export interface VimInputListener extends BaseInputListener {
  mode: 'vim'
  onModeChange: CallbackSlot<(mode: string) => void>
  onKeyPress: CallbackSlot<(key: string) => void>
  onCommandDone: CallbackSlot<() => void>
  onDispose: CallbackSlot<() => void>
}

export interface EmacsInputListener extends BaseInputListener {
  mode: 'emacs'
  onDidMarkChange: CallbackSlot<(isMarkSet: boolean) => void>
  onDidChangeKey: CallbackSlot<(key: string) => void>
}

export type InputListener = EmacsInputListener | VimInputListener
