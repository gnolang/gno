import { linter } from '@codemirror/lint'

import { getStateData } from '../state'

/**
 * Returns extension based on CM linter that renders format errors from document state.
 */
export const newFormatErrorsRenderer = () =>
  linter((view) => {
    const { diagnostics } = getStateData(view.state)
    return diagnostics ?? []
  })
