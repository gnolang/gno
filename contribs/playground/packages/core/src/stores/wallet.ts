import { BigNumber } from 'bignumber.js'
import { flow, toGenerator, types, type IAnyModelType, type Instance } from 'mobx-state-tree'
import { persist } from 'mst-persist'

import { getChainById, type Chain } from '../config'
import { DEFAULT_WALLET, supportedWallets, WalletService } from '../services'
import type { WalletAccount, WalletAPIAdapter } from '../types'
import { convertUnitToDecimal, getGnotCoin, semverCompare, truncateAddress } from '../utils'

const DEFAULT_MIN_BALANCE = 3

export const WalletStore = types
  .model('WalletStore', {
    provider: types.optional(types.string, DEFAULT_WALLET),
    hasEstablished: types.optional(types.boolean, false),
    shouldAutoConnect: types.optional(types.boolean, false),
    adapter: types.maybeNull(types.frozen<WalletAPIAdapter>()),
  })
  .volatile(() => ({
    account: undefined as WalletAccount | undefined,
    isConnecting: false,
    isInstalled: false,
    isUptodate: true,
    isSyncing: false,
  }))
  .views((self) => ({
    get state() {
      if (self.account) {
        return 'connected'
      } else if (!self.isUptodate) {
        return 'outdated'
      } else if (self.isConnecting) {
        return 'connecting'
      } else if (self.isSyncing) {
        return 'connecting'
      } else if (self.isInstalled) {
        return 'installed'
      } else {
        return 'unset'
      }
    },

    get chainDetails() {
      if (!self.account) return undefined
      return getChainById(self.account.chain)
    },

    getNetwork: async (): Promise<Chain | undefined> => {
      if (!self.account) return undefined
      const network = await self.adapter?.getNetwork()
      if (!network) {
        return undefined
      }

      return {
        id: network.id,
        displayName: network.name,
        rpcUrl: network.rpcUrl,
      } satisfies Chain
    },

    needsFunds() {
      if (!self.account) return false

      const coin = getGnotCoin(self.account.balance)
      if (!coin) return true

      const result = convertUnitToDecimal(coin.amount, coin.denom)

      return new BigNumber(result).isLessThan(DEFAULT_MIN_BALANCE)
    },

    truncatedAddress() {
      if (!self.account) return ''
      return truncateAddress(self.account.address)
    },
  }))
  .actions((self) => ({
    detect: flow(function* detect() {
      if (!self.adapter) return Promise.reject(new Error('No adapter'))

      self.isInstalled = yield self.adapter.detect()

      if (!self.isInstalled) {
        throw new Error('Wallet not detected')
      }
    }),
    checkVersion: flow(function* check() {
      if (!self.adapter) return Promise.reject(new Error('No adapter'))

      const walletVer = self.adapter.getVersion()
      self.isUptodate = !!walletVer && semverCompare(walletVer, supportedWallets[self.provider].minVer) >= 0

      if (!self.isUptodate) {
        throw new Error('Wallet outdated')
      }
    }),
    connect: flow(function* connect() {
      if (!self.adapter) return Promise.reject(new Error('No adapter'))

      self.isConnecting = true

      try {
        yield self.adapter.connect()
        self.hasEstablished = true
        self.isConnecting = false
      } catch (_error) {
        self.isConnecting = false
      }
    }),
    disconnect() {
      self.hasEstablished = false
      self.account = undefined
    },

    syncAccount: flow(function* syncAccount() {
      if (!self.adapter) return
      self.isSyncing = true
      const account = yield* toGenerator(self.adapter.getAccount())
      self.isSyncing = false
      self.account = account
      return account
    }),
  }))
  .actions((self) => ({
    /**
     * Sets arbitrary chain, supports custom networks.
     *
     * @param id Unique chain ID.
     * @param name Display name.
     * @param rpcUrl RPC URL.
     */
    setChain: flow(function* ({ id, displayName: name, rpcUrl }: Chain) {
      if (!self.adapter) return

      // Ensure the network is added to the wallet before switching
      yield self.adapter.addNetwork({ id, name, rpcUrl })
      yield self.adapter.switchNetwork(id)
      yield self.syncAccount()
    }),

    setAdapter() {
      if (!self.adapter) {
        self.adapter = WalletService.getAdapter(self.provider)
      }
    },
    configAdapter() {
      if (!self.adapter) return
      self.adapter.onChangeNetwork(() => self.syncAccount())
      self.adapter.onChangeAccount(() => self.syncAccount())
      self.account = undefined
    },
  }))
  .actions((self) => ({
    switchNetwork: flow(function* switchNetwork(network: string) {
      const chain = getChainById(network)
      if (!self.adapter || !chain) return

      yield self.setChain(chain)
    }),
  }))
  .actions((self) => ({
    start: flow(function* () {
      self.configAdapter()
      yield self.connect()
      yield self.syncAccount()
    }),
    setup: flow(function* () {
      self.setAdapter()
      yield self.detect()
      yield self.checkVersion()
    }),
  }))
  .actions((self) => ({
    afterCreate: flow(function* afterCreate() {
      const shouldAutoConnect = self.shouldAutoConnect

      // self.adapter gets lost after persist,
      // so we need to freeze it and reassign it
      const adapter = Object.freeze(self.adapter)

      yield persist('wallet', self as Instance<IAnyModelType>, { whitelist: ['hasEstablished'] })

      if (!self.adapter) {
        // reassign the adapter after persist
        self.adapter = adapter
      }

      yield self.setup()

      if (shouldAutoConnect && self.hasEstablished) {
        yield self.start()
      }
    }),
  }))

export type WalletStoreType = Instance<typeof WalletStore>
