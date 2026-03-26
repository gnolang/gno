import { Popover } from '@gnostudio/react'

import { useMutation } from '@tanstack/react-query'
import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { hstack, stack } from '@/styled-system/patterns'
import { button, popover } from '@/styled-system/recipes'

export const ConnectWalletStep: React.FC = observer(() => {
  const popoverStyles = popover({ size: 'md' })
  const store = useStore()
  const {
    mutate: connect,
    isError,
    isPending,
    error,
  } = useMutation({
    mutationFn: () => store.wallet.start(),
  })

  return (
    <>
      <Popover.Title className={popoverStyles.title}>Deploy</Popover.Title>
      <div className={stack({ gap: '4', mt: '4' })}>
        <div>
          <p>Please connect your wallet to deploy</p>
        </div>
        <div className={hstack({ justifyContent: 'flex-end' })}>
          {isError && <p className={css({ color: 'red' })}>{error?.message ?? 'Failed to connect'}</p>}

          <button
            className={cx(css({ paw: 'Click+Popover+Connect' }), button())}
            onClick={() => connect()}
            disabled={isPending}
          >
            {isPending ? 'Connecting...' : 'Connect'}
          </button>
        </div>
      </div>
    </>
  )
})
