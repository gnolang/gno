import { DEFAULT_CHAIN } from '../config'

// TODO: Re-integrate faucet when a faucet service is available for gnoland1.0
export async function requestFaucetFunds(
  _receiverAddress: string,
  chainId: string,
): Promise<null> {
  if (chainId !== DEFAULT_CHAIN) {
    return null
  }
  console.warn('Faucet not available yet for this deployment')
  return null
}
