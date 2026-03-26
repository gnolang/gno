import { useEffect } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { hstack } from '@/styled-system/patterns'
import { link } from '@/styled-system/recipes'

import { Header } from './header'

export type PageType = 'play' | 'standalone'

interface Props {
  pageType: PageType
}

const CURRENT_YEAR = new Date().getFullYear()

export const Layout: React.FC<Props> = observer(({ pageType }: Props) => {
  const store = useStore()
  const theme = store.settings.theme

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  const { pathname } = useLocation()
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [pathname])

  return (
    <main
      className={css({
        w: 'full',
        h: 'full',
        display: 'flex',
        flexDirection: 'column',
      })}
    >
      <Header type={pageType} />
      <div
        className={css({
          h: pageType === 'play' ? 'full' : 'auto',
          display: 'flex',
        })}
      >
        <Outlet />
      </div>
      <footer
        className={hstack({
          gap: '6',
          bg: 'header',
          px: '8',
          py: '1',
          fontSize: 'sm',
        })}
      >
        © 2023-{CURRENT_YEAR} NewTendermint, LLC. All rights reserved.
        <Link to="/privacy" className={link()}>
          Privacy Policy
        </Link>
        <Link to="/terms" className={link()}>
          Terms of Service
        </Link>
        <a
          href="https://forms.gle/2q7WkM2XdbMsWXwDA"
          className={cx(css({ paw: 'Click+Header+Feedback' }), link())}
          target="_blank"
          rel="noopener noreferrer"
        >
          Feedback
        </a>
      </footer>
    </main>
  )
})
