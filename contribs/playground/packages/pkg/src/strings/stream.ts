import type { StringReader, StringReaderResult, StringWriter } from './types'

export function reader(stream: ReadableStream): StringReader {
  const reader = stream.getReader()
  const decoder = new TextDecoder()

  let buf = ''

  async function read() {
    return await reader.read()
  }

  async function readAll() {
    return await read().then(async function next({ done, value }): Promise<string> {
      buf += typeof value === 'string' ? value : decoder.decode(value)
      if (done) {
        return buf
      }
      return await read().then(next)
    })
  }

  async function readLine() {
    function handle() {
      const n = buf.indexOf('\n')
      if (n !== -1) {
        const line = buf.substring(0, n)
        buf = buf.substring(n + 1)
        return line
      }

      return null
    }

    const line = handle()
    if (line !== null) {
      return { done: false, line }
    }

    return await read().then(async function next({ done, value }): Promise<StringReaderResult> {
      buf += decoder.decode(value)

      const line = handle()
      if (line !== null) {
        return { done, line }
      }

      if (done) {
        return { done, line: buf }
      }

      return await reader.read().then(next)
    })
  }

  function release() {
    reader.releaseLock()
  }

  return {
    read,
    readAll,
    readLine,
    release,
  }
}

export function writer(stream: WritableStream): StringWriter {
  const writer = stream.getWriter()
  const encoder = new TextEncoder()

  async function write(data: string) {
    await writer.ready
    await writer.write(encoder.encode(data))
  }

  async function writeLine(data: string) {
    return await write(`${data}\n`)
  }

  async function close() {
    return await writer.close()
  }

  function release() {
    writer.releaseLock()
  }

  return {
    write,
    writeLine,
    close,
    release,
  }
}
