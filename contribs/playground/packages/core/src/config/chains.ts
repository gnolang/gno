import { urls } from './urls'

export interface Chain {
  id: string
  displayName: string
  rpcUrl: string
  features?: {
    /**
     * Restrict changes to package namespace other than wallet's address or custom namespace.
     * @see {@link https://docs.gno.land/concepts/namespaces/}
     * @todo Fetch this flag from /r/sys/users.IsEnabled()
     * @default false
     */
    userNamespace: boolean
  }
  banner?: {
    message: string
    type: 'warning' | 'info'
  }
  render?: {
    name: string
    url: string
    buildUrl: (pkgPath: string) => string
  }
  link?: {
    name: string
    url: string
    buildRealmUrl: (pkgPath: string) => string
    buildTxUrl: (txHash: string) => string | undefined
  }
}

/**
 * Compares if ID and RPC URL of both chains are equal
 */
export const chainsEqual = (a: Chain, b?: Chain | null) => a.id === b?.id && a.rpcUrl === b?.rpcUrl

export const CHAINS = {
  'gnoland1.0': {
    id: 'gnoland1.0',
    displayName: 'gnoland1.0',
    rpcUrl: 'https://rpc.gno.land:443',
    features: {
      userNamespace: true,
    },
    render: {
      name: 'gno.land',
      url: urls.gnoLand,
      buildUrl(pkgPath: string) {
        return `${this.url}/${pkgPath.replace('gno.land/', '')}`
      },
    },
    link: {
      name: 'Gnoscan',
      url: 'https://gnoscan.io',
      buildRealmUrl(pkgPath: string) {
        return `${this.url}/realms/details?path=${pkgPath}`
      },
      buildTxUrl(txHash: string) {
        return `${this.url}/transactions/details?txhash=${txHash}`
      },
    },
  },
  dev: {
    id: 'dev',
    displayName: 'localhost',
    rpcUrl: 'http://127.0.0.1:26657',
  },
} satisfies Record<string, Chain>

export type ChainKey = keyof typeof CHAINS
export type ChainId = string

export const DEFAULT_CHAIN: ChainKey = 'gnoland1.0'
export const DEFAULT_CHAIN_ID: ChainId = CHAINS[DEFAULT_CHAIN].id

export const getChainById = (id: string): Chain | undefined =>
  CHAINS[id as ChainKey] ?? Object.values(CHAINS).find((chain) => chain.id === id)

export const isStandardChain = (chainId: string) => getChainById(chainId) !== undefined
