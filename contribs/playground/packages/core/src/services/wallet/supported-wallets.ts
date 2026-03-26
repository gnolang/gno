import adenaDark from '@gnoide/files/src/img/adena-logo-dark.svg'
import adenaLight from '@gnoide/files/src/img/adena-logo-light.svg'
import adenaLongDark from '@gnoide/files/src/img/adena-logo-long-dark.svg'

export interface WalletProperties {
  key: string
  name: string
  website: string
  logoDark: string
  logoLight: string
  logoLongDark?: string
  logoLongLight?: string
  minVer: string
  urls: Record<string, string>
}

export const supportedWallets: Record<string, WalletProperties> = {
  adena: {
    key: 'adena',
    name: 'Adena Wallet',
    website: 'https://adena.app/',
    logoDark: adenaDark,
    logoLight: adenaLight,
    logoLongDark: adenaLongDark,
    logoLongLight: adenaLongDark,
    minVer: '1.8.4',
    urls: {
      chromeStore: 'https://chrome.google.com/webstore/detail/adena/oefglhbffgfkcpboeackfgdagmlnihnh',
      doc: 'https://docs.adena.app/user-guide/download',
    },
  },
}

export const DEFAULT_WALLET = 'adena'
