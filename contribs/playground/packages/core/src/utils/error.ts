import type { RPCResponse } from '@gnolang/tm2-js-client'
import type { ResponseDeliverTx } from '@gnolang/tm2-js-client/bin/proto/tm2/abci'

/**
 * Shim interface to avoid explicit 'axios' installation.
 */
interface AxiosError<T> {
  status: number
  statusText: string
  name: string
  message: string
  response?: {
    data?: T
  }
}

const isAxiosError = (err: any) => {
  return err?.name === 'AxiosError'
}

/**
 * ABCIError indicates an error returned by ABCI interface for `q_eval` call.
 */
export class ABCIError extends Error {
  constructor(
    public code: number,
    public data: string,
    public message: string,
  ) {
    super(message)
  }

  /**
   * Builds ABCIError from different error if it's an AxiosError with JSON-RPC response.
   *
   * Otherwise - returns original error.
   */
  static fromError(err: unknown) {
    if (isAxiosError(err)) {
      return ABCIError.fromAxiosError(err as AxiosError<RPCResponse<unknown>>)
    }

    return err
  }

  /**
   * Constructs ABCIError if error contains a valid JSON-RPC error response.
   * @param err
   */
  static fromAxiosError(err: AxiosError<RPCResponse<unknown>>) {
    if (!err.response?.data?.error) {
      // Non-RPC response
      return err
    }

    const { code, data, message } = err.response.data.error
    return new ABCIError(code, data, message ?? err.message)
  }

  static fromResponseDeliverTx(response: ResponseDeliverTx) {
    if (!response.response_base?.error) {
      return new Error('Unable to parse deliver response')
    }

    const { error, log } = response.response_base
    const reason = new TextDecoder().decode(error.value).trim()

    return new ABCIError(0, reason, log)
  }
}
