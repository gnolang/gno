import { type TerminalStore } from '@gnostudio/core'

import { createContext } from '../../create-context'

export interface TerminalContext {
  store: TerminalStore
  onClose?: () => void
}

export const [TerminalProvider, useTerminalContext] = createContext<TerminalContext>({
  name: 'TerminalContext',
  hookName: 'useTerminalContext',
  providerName: '<TerminalProvider />',
  strict: true,
})
