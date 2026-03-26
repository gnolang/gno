const decoder = new TextDecoder()

type ReadCallback = (err: any, n: number, buff?: Uint8Array) => void

export interface OutputHandler {
  write: (chunk: Uint8Array) => void
  close?: () => void
}

export interface InputHandler {
  read: (buf: Uint8Array, cb: ReadCallback) => void
  close?: () => void
}

export interface StreamOutputHandler extends OutputHandler {
  stream: ReadableStream
}

export interface StreamInputHandler extends InputHandler {
  stream: WritableStream
}

/**
 * Creates output handler for stdout and stderr that uses ReadableStream transport.
 *
 * Allows capturing and processing process output.
 *
 * Uses StreamDefaultController instead of byte streams to support Safari.
 *
 * @see https://github.com/gnostudio/studio/issues/462
 */
export const newStreamOutputHandler = (): StreamOutputHandler => {
  const chunks: Uint8Array[] = []
  let exited = false
  let controller: ReadableStreamDefaultController | null = null
  let readResolver: (() => void) | null = null

  const handle = () => {
    if (!readResolver) {
      return
    }

    if (!chunks.length) {
      if (exited) {
        controller?.close()
        readResolver()
      }
      return
    }

    chunks.forEach((chunk) => {
      controller?.enqueue(chunk)
    })
    readResolver()

    chunks.length = 0
    controller = null
    readResolver = null
  }

  const stream = new ReadableStream({
    async pull(ctrl: ReadableStreamDefaultController) {
      return await new Promise<void>((resolve) => {
        controller = ctrl
        readResolver = resolve
        handle()
      })
    },
  })

  const write = (chunk: Uint8Array) => {
    chunks.push(chunk)
    handle()
  }

  const close = () => {
    exited = true
    controller?.close()
  }

  return {
    stream,
    write,
    close,
  }
}

const concatBuffers = (a: Uint8Array, b: Uint8Array) => {
  const newBuff = new Uint8Array(a.length + b.length)
  newBuff.set(a, 0)
  newBuff.set(b, a.byteLength)
  return newBuff
}

/**
 * Creates standard output handler with MessagePort transport for web worker.
 */
export const newMessagePortOutputHandler = (port: MessagePort, isStderr = false): OutputHandler => {
  let isClosed = false
  return {
    write(chunk) {
      // should be in sync with ChannelTerminalAdapter in @gnostudio/react
      if (!isClosed) {
        const { buffer } = chunk
        port.postMessage({ payload: chunk, isStderr }, [buffer])
      }
    },
    close: () => {
      isClosed = true
    },
  }
}

/**
 * Creates standard input handler with MessagePort transport for web worker.
 */
export const newMessagePortInputHandler = (port: MessagePort): InputHandler => {
  let isClosed = false
  let unreadQueue = new Uint8Array()
  let pendingRequest: {
    buf: Uint8Array
    callback: ReadCallback
  } | null = null

  const flushQueue = () => {
    if (!unreadQueue.byteLength || !pendingRequest) {
      return
    }

    // slice operations already bounds-safe and passed length will be capped.
    const { buf, callback } = pendingRequest
    const chunk = unreadQueue.slice(0, buf.byteLength)

    buf.set(chunk)
    unreadQueue = unreadQueue.slice(buf.byteLength)

    callback(null, chunk.byteLength, buf)
  }

  port.onmessage = ({ data }: MessageEvent<ArrayBuffer>) => {
    if (isClosed) {
      return
    }

    // put data into unread queue if no one listening atm.
    unreadQueue = concatBuffers(unreadQueue, new Uint8Array(data))
    flushQueue()
  }

  return {
    read: (buf: Uint8Array, callback: ReadCallback) => {
      if (isClosed) {
        // EOF
        callback(null, 0)
        return
      }

      pendingRequest = { buf, callback }
      flushQueue()
    },
    close: () => {
      isClosed = true
      flushQueue()
      pendingRequest = null
      unreadQueue = unreadQueue.slice(0, 0)
    },
  }
}

/**
 * Creates input handler that reads from a static bytes source.
 * Returned stream is read-only and can't be written to.
 *
 * Useful for piping predefined input into stdin of a process.
 *
 * Works similar to shell pipes:
 *
 * ```
 * cat source | wasm-program
 * ```
 *
 * @param input Source data
 */
export const newInputFromData = (source: Uint8Array | string): InputHandler => {
  const encoder = new TextEncoder()
  const inputBuf = typeof source === 'string' ? encoder.encode(source) : source

  let offset = 0
  const read = (buf: Uint8Array, cb: ReadCallback) => {
    const n = Math.min(buf.length, inputBuf.length - offset)
    buf.set(inputBuf.slice(offset, offset + n))
    offset += n
    cb(null, n, buf)
  }

  return {
    read,
  }
}

/**
 * OutputRecorder is stdout/stderr handler which captures all output
 * into a buffer to be processed after process finish.
 */
export class OutputRecorder implements OutputHandler {
  private buf: ArrayBuffer
  private len = 0

  /**
   * Raw bytes content written into a buffer.
   */
  get bytes() {
    return new Uint8Array(this.buf, 0, this.len)
  }

  /**
   * Decoded string representation of content written into a buffer.
   */
  get string() {
    return decoder.decode(this.bytes)
  }

  constructor(initCapacity = 128) {
    this.buf = new ArrayBuffer(initCapacity)
  }

  #grow(newLen: number) {
    if (this.buf.byteLength >= newLen) {
      return
    }

    const newCap = Math.round(newLen * 1.5)
    const newBuf = new ArrayBuffer(newCap)
    new Uint8Array(newBuf).set(new Uint8Array(this.buf))
    this.buf = newBuf
  }

  write(src: Uint8Array) {
    const newLen = this.len + src.byteLength
    this.#grow(newLen)
    new Uint8Array(this.buf).set(src, this.len)
    this.len += src.byteLength
  }

  close() {
    // no-op
  }
}

/**
 * Returns output handler that writes output to console.log.
 *
 * Used when no there is no need to handle output.
 */
export const newLogOutputWriter = (isError = false): OutputHandler => {
  let outputBuf = ''

  const writeFn = isError ? console.error : console.log

  const write = (buf: Uint8Array) => {
    outputBuf += decoder.decode(buf)
    const nl = outputBuf.lastIndexOf('\n')
    if (nl !== -1) {
      writeFn(outputBuf.substring(0, nl))
      outputBuf = outputBuf.substring(nl + 1)
    }
  }

  const close = () => {
    if (outputBuf.length) {
      console.log(outputBuf)
    }
  }

  return {
    write,
    close,
  }
}

/**
 * Creates a no-op stdin handler.
 *
 * Used when no input processing is required.
 * @returns
 */
export const newNoopInputHandler = (): InputHandler => {
  const read = (_buf: Uint8Array, cb: ReadCallback) => {
    cb(null, 0)
  }

  return {
    read,
  }
}
