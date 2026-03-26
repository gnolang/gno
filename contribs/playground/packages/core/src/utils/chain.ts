import { JSONRPCProvider } from '@gnolang/tm2-js-client'

/**
 * Fetches chain ID of a node using RPC URL.
 * @param rpcUrl RPC URL.
 */
export const discoverChainId = async (rpcUrl: string) => {
  const provider = new JSONRPCProvider(rpcUrl)
  const {
    node_info: { network },
  } = await provider.getStatus()
  return network
}
