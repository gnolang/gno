import { createContext, useContext } from 'react'

import { type Instance } from 'mobx-state-tree'

import { type rootStore } from '../store'

type Store = Instance<typeof rootStore>
const StoreContext = createContext<Instance<Store>>(null as unknown as Store)

export const StoreProvider = StoreContext.Provider

export function useStore() {
  return useContext(StoreContext)
}
