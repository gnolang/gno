import { prepareVMABCIEvaluateExpressionQuery, VMEndpoint } from '@gnolang/gno-js-client'
import { ABCIEndpoint, newRequest, RestService, type ABCIResponse } from '@gnolang/tm2-js-client'

import { ABCIError } from '../../utils/error'

interface EvalExprParams {
  rpcUrl: string
  pkgPath: string
  expr: string
  chainId: string
}

/**
 * Evaluates a function expression in read-only mode on GnoVM without transaction.
 *
 * @see GnoJSONRPCProvider.evaluateExpression
 */
export const evalExpr = async ({ rpcUrl, pkgPath, expr }: EvalExprParams) => {
  try {
    // GnoJSONRPCProvider doesn't return raw ABCI response.
    // Copy-pasta from https://github.com/gnolang/gno-js-client/blob/main/src/provider/jsonrpc/jsonrpc.ts#L25
    //
    // See: https://github.com/gnolang/gno-js-client/issues/114
    const query = prepareVMABCIEvaluateExpressionQuery([pkgPath, expr])

    const result = await RestService.post<ABCIResponse>(rpcUrl, {
      request: newRequest(ABCIEndpoint.ABCI_QUERY, [`vm/${VMEndpoint.EVALUATE}`, query, '0', false]),
    })

    return result
  } catch (err) {
    // Extract JSON-RPC error response from Axios error (if any).
    throw ABCIError.fromError(err)
  }
}
