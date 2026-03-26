import { CHAINS, chainsEqual, DEFAULT_CHAIN, getChainById, type Chain } from '@gnostudio/core'

import { types } from 'mobx-state-tree'
import { persist } from 'mst-persist'

const ChainModel = types.model({
  id: types.string,
  displayName: types.string,
  rpcUrl: types.string,
})

export const Chains = types
  .model({
    selectedChain: types.maybeNull(ChainModel),
    customChains: types.optional(types.array(ChainModel), []),
  })
  .views((self) => ({
    isSelected: (c: Chain) => chainsEqual(c, self.selectedChain),
  }))
  .actions((self) => ({
    setActiveChain(chain: Chain) {
      self.selectedChain = ChainModel.create(chain)
    },
    addChain(chain: Chain) {
      self.customChains.push(chain)
    },
    removeChainByIndex(i: number) {
      const chain = self.customChains[i]
      if (!chain) {
        return
      }

      if (chainsEqual(chain, self.selectedChain)) {
        self.selectedChain = CHAINS[DEFAULT_CHAIN]
      }
      self.customChains.splice(i, 1)
    },
  }))
  .actions((self) => ({
    afterCreate() {
      if (!self.selectedChain) {
        self.setActiveChain(CHAINS[DEFAULT_CHAIN])
      }

      persist('chains', self, {
        // Do not persist selected chain as it might be out of sync with wallet.
        whitelist: ['customChains'],
      }).catch(console.error)
    },
  }))
  .actions((self) => ({
    /**
     * Sets active chain by chain ID.
     *
     * Tries to find chain in default chains, then in custom chains.
     * Sets a stub chain value if chain is not registered.
     *
     * Used to sync active wallet chain from account information
     * with only chain ID available.
     *
     * @param chainId Chain ID.
     */
    setActiveChainById(chainId: string) {
      const builtin = getChainById(chainId)
      if (builtin) {
        self.setActiveChain(builtin)
        return
      }

      for (const chain of self.customChains) {
        if (chain.id === chainId) {
          self.selectedChain = chain
          return
        }
      }

      self.selectedChain = CHAINS[DEFAULT_CHAIN]
    },
  }))
