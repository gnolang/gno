import React, { useState } from 'react'
import { Link } from 'react-router-dom'

import GnolandLogo from '@gnoide/files/src/img/gnoland-logo-gnome.svg?react'
import { CHAINS, DEFAULT_CHAIN } from '@gnostudio/core'

import { css, cx } from '@/styled-system/css'
import { divider, hstack, spacer, visuallyHidden } from '@/styled-system/patterns'

import { DeployPopover } from '../deploy-popover'
import { ExamplesPopover } from '../examples-popover'
import { ImportPopover } from '../import-popover'
import { type PageType } from '../layout'
import { SharePopover } from '../share-popover'
import { HeaderActions } from './header-actions'
import { HeaderInfo } from './header-info'
import { HeaderSettings } from './header-settings'

interface Props {
  type?: PageType
}

const HeaderMenu: React.FC = () => {
  const verticalDividerClass = divider({ orientation: 'vertical', h: '4' })

  return (
    <div className={hstack({ gap: '5' })}>
      <nav>
        <ul className={hstack({ gap: '4' })}>
          <li>
            <ExamplesPopover />
          </li>
          <li>
            <hr className={verticalDividerClass} />
          </li>
          <li>
            <ImportPopover />
          </li>
          <li>
            <SharePopover />
          </li>
          <li>
            <DeployPopover />
          </li>
        </ul>
      </nav>
      <HeaderActions />
    </div>
  )
}

const HeaderLogo: React.FC = () => {
  return (
    <Link to="/" className={cx(css({ paw: 'Click+Header+Home' }), css({ flexShrink: '0' }))}>
      <GnolandLogo className={css({ color: 'current' })} />
    </Link>
  )
}

export const Header: React.FC<Props> = ({ type = 'play' }: Props) => {
  const [isBannerVisible, setIsBannerVisible] = useState(true)
  const clearInfo = () => setIsBannerVisible(false)

  return (
    <div>
      {isBannerVisible && (
        <HeaderInfo clearInfo={clearInfo}>
          <strong
            className={css({
              textAlign: 'center',
              display: 'block',
              color: 'inherit',
              margin: '0 auto',
            })}
          >
            Testnet <span className={css({ fontWeight: 'extrabold' })}>{CHAINS[DEFAULT_CHAIN].displayName}</span> is
            live. Switch networks or add a custom RPC endpoint anytime.
          </strong>
        </HeaderInfo>
      )}
      <header
        data-testid="app-header"
        className={hstack({
          gap: '8',
          bg: 'header',
          h: '65px',
          px: '8',
          py: '4',
        })}
      >
        {type === 'play' && (
          <>
            <h1>
              <span className={visuallyHidden()}>Gno Playground</span>
              <HeaderLogo />
            </h1>
            <HeaderMenu />
          </>
        )}
        {type === 'standalone' && <HeaderLogo />}
        <div className={spacer()} />
        <HeaderSettings type={type} />
      </header>
    </div>
  )
}
