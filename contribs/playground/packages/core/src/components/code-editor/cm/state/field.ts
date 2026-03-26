import { StateEffect, StateField } from '@codemirror/state'

import { defaultStateData, type StateData } from './types'

/**
 * Effect to update per-view state data.
 *
 * @see StateData
 */
export const updateStateDataEffect = StateEffect.define<Partial<StateData>>()

/**
 * Field to store and update StateData in EditorState.
 */
export const stateDataField = StateField.define<StateData>({
  create() {
    return defaultStateData
  },
  update(currentValue, transaction) {
    for (const effect of transaction.effects) {
      if (effect.is(updateStateDataEffect)) {
        return { ...currentValue, initialized: true, ...effect.value }
      }
    }

    return currentValue
  },
})
