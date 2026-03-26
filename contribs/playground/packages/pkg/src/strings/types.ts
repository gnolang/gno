export interface Releaser {
  release: () => void
}

export interface Reader {
  read: () => Promise<ReadableStreamReadResult<BufferSource>>
}

export interface StringReaderResult {
  done: boolean
  line: string
}

export interface StringReader extends Reader, Releaser {
  readAll: () => Promise<string>
  readLine: () => Promise<StringReaderResult>
}

export interface StringWriter extends Releaser {
  write: (data: string) => Promise<void>
  writeLine: (data: string) => Promise<void>
  close: () => Promise<void>
}
