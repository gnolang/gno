import { supportedWallets } from '@gnostudio/core'
import { Popover } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { hstack, stack } from '@/styled-system/patterns'
import { button, link, popover } from '@/styled-system/recipes'

interface Props {
  error: 'update' | 'install'
}

export const InstallWalletStep: React.FC<Props> = observer((props: Props) => {
  const store = useStore()
  const wallet = store.wallet.provider
  const walletInfo = supportedWallets[wallet]

  const popoverStyles = popover({ size: 'md' })

  const content = {
    update: {
      title: `Update ${walletInfo.name}`,
      desc: `Update ${walletInfo.name} Wallet extension and reload the page to deploy.`,
    },
    install: {
      title: `Install ${walletInfo.name}`,
      desc: `Install ${walletInfo.name} Wallet extension and reload the page to deploy.`,
    },
  }

  return (
    <>
      <Popover.Title className={popoverStyles.title}>{content[props.error].title}</Popover.Title>
      <div className={stack({ gap: '4', mt: '4' })}>
        <div>
          <p>{content[props.error].desc}</p>
          <a
            href={walletInfo.urls.chromeStore}
            title={`Install ${walletInfo.name}`}
            className={link()}
            target="_blank"
            rel="noreferrer"
          >
            {walletInfo.name} in Chrome Web Store
          </a>
        </div>
        <div className={hstack({ justifyContent: 'flex-end' })}>
          <button className={button()} onClick={() => location.reload()}>
            Reload the page
          </button>
        </div>
      </div>
    </>
  )
})
