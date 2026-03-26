import React, { useState } from 'react'

import { CHAINS, DEFAULT_CHAIN, discoverChainId, supportedWallets, type Chain } from '@gnostudio/core'

import { observer } from 'mobx-react-lite'
import { getSnapshot } from 'mobx-state-tree'

import { useStore } from '@/contexts'

import { NetworkSelector } from './network-selector'

const standardChains = [CHAINS[DEFAULT_CHAIN], CHAINS.dev].filter(Boolean)

const displayError = (prefix: string, err: any) => {
  const msg: string = err?.message?.toString() ?? err.toString()
  window.alert(`${prefix}: ${msg}`)
}

/**
 * ConnectedNetworkSelector is a NetworkSelector connected to the store.
 */
export const ConnectedNetworkSelector: React.FC = observer(() => {
  const store = useStore()
  const [isBusy, setBusy] = useState(false)

  // Standard chains are kept out of store to avoid storing outdated chains.
  const customChains = getSnapshot(store.chains.customChains)

  const onChainAdd = (c: Chain) => {
    store.chains.addChain(c)
  }

  const onConnectRequest = async () => {
    setBusy(true)
    try {
      await store.wallet.start()
    } catch (err) {
      displayError('Failed to connect to the wallet', err)
    } finally {
      setBusy(false)
    }
  }

  const resolveChainId = async (rpcUrl: string) => {
    try {
      const cid = await discoverChainId(rpcUrl)
      if (!cid?.length) {
        window.alert('Address is not a valid node')
        return null
      }
      return cid
    } catch (err: any) {
      displayError('Failed to obtain chain ID', err)
      return null
    }
  }

  const onChainSelect = async (c: Chain) => {
    if (!store.wallet.isInstalled) {
      store.chains.setActiveChain(c)
      return
    }

    setBusy(true)
    try {
      await store.wallet.setChain(c)
      store.chains.setActiveChain(c)
    } catch (err) {
      displayError('Failed to set chain', err)
    } finally {
      setBusy(false)
    }
  }

  const onChainRemove = async (c: Chain, i: number) => {
    if (store.chains.isSelected(c)) {
      return
    }

    store.chains.removeChainByIndex(i)
  }

  return (
    <NetworkSelector
      predefinedItems={standardChains}
      editableItems={customChains}
      selectedItem={store.chains.selectedChain}
      disabled={isBusy}
      onAdd={onChainAdd}
      onChange={onChainSelect}
      onRemove={onChainRemove}
      chainIdProvider={resolveChainId}
      connected={!!store.wallet.account}
      onConnectRequest={onConnectRequest}
      walletAvailability={{
        isInstalled: store.wallet.isInstalled,
        walletName: supportedWallets.adena.name,
        installPrompt: `You need to install ${supportedWallets.adena.name} and reload the page to interact with realms. Make sure that you have at least one account created in your wallet.`,
        installUrl: supportedWallets.adena.urls.chromeStore,
      }}
    />
  )
})
