import React from 'react'

import { stack } from '@/styled-system/patterns'
import { button, text } from '@/styled-system/recipes'

interface Props {
  onConnectRequest?: () => void
  disabled?: boolean
  walletAvailability: {
    isInstalled: boolean
    walletName: string
    installPrompt: string
    installUrl: string
  }
}

export const WalletConnectPrompt: React.FC<Props> = ({
  onConnectRequest,
  disabled,
  walletAvailability: { isInstalled, installPrompt, installUrl, walletName },
}) => (
  <div
    className={stack({ gap: '4', alignItems: 'center', justifyContent: 'center', pr: '2', pl: '2', pt: '3', pb: '3' })}
    data-testid="wallet-connect-prompt"
  >
    <span className={text({ align: 'center' })}>
      {isInstalled ? 'Please connect your wallet to select a network.' : installPrompt}
    </span>
    {isInstalled ? (
      <button type="button" className={button()} disabled={disabled} onClick={onConnectRequest}>
        Connect Wallet
      </button>
    ) : (
      <div className={stack({ gap: '1', flex: '1', width: '100%' })}>
        <a href={installUrl} className={button({ block: true })} target="_blank" rel="noreferrer noopener">
          Install {walletName}
        </a>
        <button
          type="button"
          className={button({ block: true, variant: 'outline' })}
          onClick={() => window.location.reload()}
        >
          Reload page
        </button>
      </div>
    )}
  </div>
)
