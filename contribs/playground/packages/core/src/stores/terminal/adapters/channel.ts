import { type TerminalAdapter, type Writer } from './types'

const encoder = new TextEncoder()

interface OutputHandlerMessage {
  payload: ArrayBuffer
  isStderr: boolean
}

/**
 * Terminal adapter for communication over MessageChannel.
 */
export class ChannelTerminalAdapter implements TerminalAdapter {
  constructor(
    private readonly port: MessagePort,
    private readonly isReadOnly = false,
  ) {
    // stub
  }

  async write(text: string) {
    if (!this.isReadOnly) {
      const { buffer } = encoder.encode(text)
      this.port.postMessage(buffer, [buffer])
    }
  }

  async readInto(writer: Writer) {
    this.port.onmessage = ({ data }) => {
      // See: newMessagePortOutputHandler in @gnostudio/wasm package.
      const { payload } = data as OutputHandlerMessage
      writer.write(new Uint8Array(payload))
    }
  }

  dispose() {
    this.port.close()
  }
}
