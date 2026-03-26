import {
  broadcastTransaction,
  getChainById,
  getUserNamespace,
  injectStudioMetaFiles,
  isValidPkgPath,
  packagePaths,
  parsePkgPath,
  requestFaucetFunds,
  TransactionBuilder,
  type WalletStoreType,
} from '@gnostudio/core'
import { detectModuleName } from '@gnostudio/wasm'

import { reaction, when } from 'mobx'
import { addDisposer, flow, getRoot, getSnapshot, toGenerator, types, type IDisposer } from 'mobx-state-tree'

import { type RootStore } from './root'

export type Step = 'connect' | 'deploy' | 'install' | 'update'

export const Deployer = types
  .model({
    pathPrefix: types.optional(types.string, packagePaths.pathPrefix),
    pathType: types.optional(types.enumeration(['r', 'p']), 'r'),
    pathPart: types.optional(types.string, ''),
    pathNamespace: types.optional(types.string, ''),
  })
  .volatile(() => ({
    step: 'install' as Step,
    broadcastResult: undefined as undefined | Awaited<ReturnType<typeof broadcastTransaction>>,
    userNamespace: '',
  }))
  .views((self) => ({
    get pkgPath() {
      return (
        [self.pathPrefix, self.pathType, self.pathNamespace, self.pathPart]
          // Remove leading and trailing slashes
          .map((part) => part.replace(/^\/|\/$/g, '').trim())
          .join('/')
      )
    },

    get pkgName() {
      const chunks = self.pathPart.split('/')
      return chunks[chunks.length - 1]
    },

    get hasValidPath() {
      return isValidPkgPath(this.pkgPath)
    },

    get txHash() {
      if (!self.broadcastResult) return undefined

      const { hash } = self.broadcastResult

      return hash
    },

    wallet(): WalletStoreType {
      return getRoot<RootStore>(self).wallet
    },
  }))
  .views((self) => ({
    get canModifyNamespace() {
      const chain = self.wallet().chainDetails
      if (!chain) return false

      const isNamespaceFeatureEnabled = chain?.features?.userNamespace ?? false

      return !isNamespaceFeatureEnabled
    },
  }))
  .actions((self) => ({
    setStep(step: typeof self.step) {
      self.step = step
    },

    setPathType(type: string) {
      self.pathType = type
    },

    setPathPart(pkgName: string) {
      self.pathPart = pkgName
    },

    setPathNamespace(namespace: string) {
      self.pathNamespace = namespace
    },
  }))
  .actions((self) => ({
    extractPathFromWorkspace() {
      const root = getRoot<RootStore>(self)
      const pkgPath = detectModuleName(getSnapshot(root.workbench.files))
      if (!pkgPath) return

      const { type, path } = parsePkgPath(pkgPath)
      if (type) self.setPathType(type)
      if (path) self.setPathPart(path)
    },

    extractNamespaceFromWallet: flow(function* extractNamespaceFromWallet() {
      const account = self.wallet().account
      const rpcUrl = getRoot<RootStore>(self).chains.selectedChain?.rpcUrl as string

      // Reset the namespace if the wallet changes
      self.setPathNamespace(self.canModifyNamespace ? '' : (account?.address ?? ''))
      self.userNamespace = ''

      if (!account) return

      // Try to get the namespace from the wallet
      const user = yield* toGenerator(
        getUserNamespace({
          address: account.address,
          chainId: account.chain,
          rpcUrl,
        }),
      )

      if (user) self.userNamespace = user
    }),
  }))
  .actions((self) => ({
    connect() {
      self.extractPathFromWorkspace()
      self.setStep('deploy')
    },

    deploy: flow(function* deploy(targetChain: string) {
      let account = self.wallet().account
      const adapter = self.wallet().adapter

      if (!adapter) return

      const root = getRoot<RootStore>(self)
      const files = root.workbench.tabs

      // Sync account to get the latest balance
      account = yield self.wallet().syncAccount()
      if (!account) return

      // Make sure the wallet is connected to the right network
      if (targetChain !== account.chain) {
        yield self.wallet().switchNetwork(targetChain)
      }

      if (root.wallet.needsFunds()) {
        console.log(`Requesting funds for ${account.address}...`)
        try {
          yield requestFaucetFunds(account.address, targetChain)
        } catch (error) {
          console.error('Failed to request funds:', error)
        }
      }

      const transactionBuilder = new TransactionBuilder()

      transactionBuilder.addPkg({
        creator: account.address,
        deposit: '',
        data: {
          name: self.pkgName,
          path: self.pkgPath,
          files: injectStudioMetaFiles(files, self.pkgPath)
            .map((file) => ({
              name: file.path,
              body: file.content,
            }))
            .sort((a, b) => a.name.localeCompare(b.name)),
        },
      })

      transactionBuilder.setGas({ gasFee: 50000, gasWanted: 1e7 })

      const transaction = transactionBuilder.build()
      const result = yield adapter.signTransaction(transaction)

      if (result?.status === 'failure') {
        console.error(result)
        throw new Error(result.message)
      }

      // Handle cases when user has a custom Adena network not known to Playground (issue: #1364)
      const chainId = self.wallet().account?.chain as string
      let chain = getChainById(chainId)
      if (!chain) {
        chain = yield self.wallet().getNetwork()
      }

      const response = yield broadcastTransaction(result.data?.encodedTransaction as string, chainId, chain?.rpcUrl)
      self.broadcastResult = response
      root.projects.saveDeployedProject({ id: self.pkgPath })

      return result
    }),

    afterCreate() {
      const disposers: IDisposer[] = []

      // Move to connect step if the wallet is disconnected
      disposers.push(
        reaction(
          () => self.wallet().state,
          (state) => {
            const stepper = {
              connected: 'deploy',
              outdated: 'update',
              installed: 'connect',
              connecting: 'connect',
              unset: 'install',
            }
            self.setStep(stepper[state] as Step)
          },
        ),
      )

      // Extract the package path from the workspace only once
      disposers.push(
        when(
          () => self.step === 'deploy',
          () => {
            self.extractPathFromWorkspace()
          },
        ),
      )

      // Extract the namespace from the wallet whenever the account changes
      disposers.push(
        reaction(
          () => self.wallet().account,
          (account, prev) => {
            // Skip if the account hasn't changed
            if (account?.address === prev?.address && account?.chain === prev?.chain) return
            self.extractNamespaceFromWallet().catch(console.error)
          },
        ),
      )

      // Dispose all reactions when the store is destroyed
      disposers.forEach((disposer) => addDisposer(self, disposer))
    },
  }))
