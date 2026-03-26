import { ERR_UNKNOWN, JSONRPCError } from '@open-rpc/client-js/build/Error'
import { getBatchRequests, getNotifications, type JSONRPCRequestData } from '@open-rpc/client-js/build/Request'
import { Transport } from '@open-rpc/client-js/build/transports/Transport'

type LazyMessagePort = () => Promise<MessagePort>

interface LifecycleHooks {
  /**
   * Called before connection is established.
   */
  onConnect?: () => void

  /**
   * Called before transport is closed.
   */
  onClose?: () => void
}

const isLazyMessagePort = (port: MessagePort | LazyMessagePort): port is LazyMessagePort => typeof port === 'function'

/**
 * Implements JSON-RPC transport via MessagePort.
 *
 * Based on WebSocketTransport from open-rpc lib.
 */
export class MessagePortTransport extends Transport {
  private readonly portProvider?: LazyMessagePort
  private portPromise?: Promise<MessagePort>

  /**
   * Transport constructor.
   *
   * Port can be a MessagePort or a async function which can be used to create a port
   * later, right before connect.
   *
   * @param source Message port or a lazy message port provider.
   * @param hooks Transport lifecycle hooks.
   */
  constructor(
    source: MessagePort | LazyMessagePort,
    private readonly hooks?: LifecycleHooks,
  ) {
    super()

    if (isLazyMessagePort(source)) {
      this.portProvider = source
    } else {
      this.portPromise = Promise.resolve(source)
    }
  }

  private async getPort() {
    if (!this.portPromise) {
      if (!this.portProvider) {
        // This never should happen as either one of two always exist in ctor.
        throw new Error('MessagePort provider is not configured.')
      }

      // Call message port provider at first boot.
      this.portPromise = this.portProvider()
    }

    return await this.portPromise
  }

  async connect() {
    const port = await this.getPort()
    port.onmessage = ({ data }) => {
      this.transportRequestManager.resolveResponse(data)
    }
  }

  close() {
    this.portPromise = undefined
    this.hooks?.onClose?.()
  }

  async sendData(data: JSONRPCRequestData, timeout: number | null = null): Promise<any> {
    const promise = this.transportRequestManager.addRequest(data, timeout)
    const notifications = getNotifications(data)

    try {
      const port = await this.getPort()
      const request = this.parseData(data)
      port.postMessage(request)
      this.transportRequestManager.settlePendingRequest(notifications)
    } catch (err) {
      const jsonError = new JSONRPCError((err as any).message, ERR_UNKNOWN, err)

      this.transportRequestManager.settlePendingRequest(notifications, jsonError)
      this.transportRequestManager.settlePendingRequest(getBatchRequests(data), jsonError)

      throw jsonError
    }

    return await promise
  }
}
