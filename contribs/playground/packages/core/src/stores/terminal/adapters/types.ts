export interface Writer {
  write: (text: string | Uint8Array) => void
}

/**
 * TerminalAdapter is Terminal MST adapter to support different stdio sources.
 */
export interface TerminalAdapter {
  /**
   * Writes a given chunk of text into source.
   */
  write: (text: string) => Promise<void>

  /**
   * Attaches a writer to a adapter to send incoming characters into a terminal.
   */
  readInto: (writer: Writer) => Promise<void>

  /**
   * Disposes adapter, optional.
   */
  dispose?: () => void
}
