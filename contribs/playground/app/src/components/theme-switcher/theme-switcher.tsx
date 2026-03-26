import React from 'react'
import { PiMoonFill, PiSunFill } from 'react-icons/pi'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { button } from '@/styled-system/recipes'

export const ThemeSwitcher: React.FC = observer(() => {
  const store = useStore()

  const label = store.settings.isDark ? 'Switch to light theme' : 'Switch to dark theme'
  const otherScheme = store.settings.isDark ? 'light' : 'dark'

  const handleClick = () => {
    store.settings.setTheme(otherScheme)
  }

  return (
    <button
      title={label}
      aria-label={label}
      onClick={handleClick}
      className={cx(css({ paw: 'Click+Header+Home' }), button({ variant: 'ghost' }))}
    >
      {store.settings.isDark ? <PiSunFill size={24} /> : <PiMoonFill size={24} />}
    </button>
  )
})
