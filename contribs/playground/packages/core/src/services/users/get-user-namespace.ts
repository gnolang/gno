import { evalExpr } from '../transaction'

interface GetWalletNamespaceParams {
  address: string
  rpcUrl: string
  chainId: string
}

const usersRealmPath = 'gno.land/r/demo/users'

/**
 * Extract the namespace of a wallet by address,
 * we use the users realm to check if the address is a user
 */
export async function getUserNamespace({
  address,
  rpcUrl,
  chainId,
}: GetWalletNamespaceParams): Promise<string | undefined> {
  try {
    const { response } = await evalExpr({
      rpcUrl,
      pkgPath: usersRealmPath,
      expr: `Render("${address}")`,
      chainId,
    })

    // Expected output:
    // ("## user gnoland\n\n * address = g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7\n * 0 invites\n * invited by g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq\n\n\n" string)
    const result = atob(response.ResponseBase.Data ?? '')
    const match = result.match(/user\s([a-zA-Z0-9_]+)/)

    return match?.[1]
  } catch (err) {
    console.error(err)
  }
}
