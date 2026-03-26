import { Compartment, type Extension } from '@codemirror/state'
import { keymap } from '@codemirror/view'
import { emacs } from '@replit/codemirror-emacs'
import { vim } from '@replit/codemirror-vim'
import { vscodeKeymap } from '@replit/codemirror-vscode-keymap'

import type { InputMode } from '../../common'

const getDefaultInputExtension = () => keymap.of(vscodeKeymap)

/**
 * Loads and returns an extension by input mode.
 *
 * Extensions are loaded on demand.
 */
const getExtensionByMode = (mode?: InputMode): Extension => {
  // TODO: load modules dynamically
  switch (mode) {
    case 'emacs': {
      return emacs()
    }
    case 'vim': {
      return vim()
    }
    default:
      return getDefaultInputExtension()
  }
}

export const inputModeCompartment = new Compartment()

/**
 * Returns a new compartment for input mode extension with initial extension based on input mode.
 *
 * As extensions are loaded on demand, unlike other compartment constructors this function is async.
 */
export const newInputModeCompartment = (mode?: InputMode) => inputModeCompartment.of(getExtensionByMode(mode))

/**
 * Returns a new state effect to reconfigure input mode.
 *
 * Asynchronous as extensions are loaded on demand.
 */
export const updateInputModeEffect = (mode?: InputMode) => inputModeCompartment.reconfigure(getExtensionByMode(mode))

/**
 * Returns a state effect that resets input mode back to regular.
 */
export const resetInputModeEffect = () => inputModeCompartment.reconfigure(getDefaultInputExtension())
