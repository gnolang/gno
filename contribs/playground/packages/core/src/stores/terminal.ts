import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import elegantSpinner from 'elegant-spinner'
import { makeAutoObservable } from 'mobx'

import { ChannelTerminalAdapter, type TerminalAdapter } from './terminal/adapters'

const spinner = elegantSpinner()

const ANSI_LEFT = '\x1B[1D'
const ANSI_CHAR_DELETE = '\x1B[1P'
const ANSI_LEFT_CHAR_DELETE = `${ANSI_LEFT}${ANSI_CHAR_DELETE}`

export class TerminalStore {
  protected xterm: Terminal | undefined
  protected fitAddon: FitAddon | undefined
  protected spinnerInterval: NodeJS.Timeout | undefined
  protected adapter: TerminalAdapter | undefined

  public isOpen = false

  constructor() {
    makeAutoObservable(this)
  }

  public startSpinner() {
    this.spinnerInterval = setInterval(() => {
      this.xterm?.write(ANSI_LEFT)
      this.xterm?.write(spinner())
    }, 100)
  }

  public stopSpinner() {
    clearInterval(this.spinnerInterval)
    this.xterm?.write(ANSI_LEFT)
    this.spinnerInterval = undefined
  }

  public attachToMessagePort(port: MessagePort, readOnly = false) {
    return this.setAdapter(new ChannelTerminalAdapter(port, readOnly))
  }

  public open() {
    this.isOpen = true
  }

  public close() {
    this.isOpen = false
  }

  public toggle() {
    this.isOpen = !this.isOpen
  }

  public write(text: string) {
    this.xterm?.write(text)
  }

  public clear() {
    this.xterm?.reset()
  }

  public fit() {
    this.fitAddon?.fit()
  }

  public async setAdapter(adapter: TerminalAdapter) {
    this.adapter?.dispose?.()
    this.adapter = adapter
    await adapter.readInto({
      write: (text) => {
        this.stopSpinner()
        this.xterm?.write(text)
      },
    })
  }

  public printFatalError(err: Error) {
    this.adapter?.dispose?.()
    this.xterm?.writeln(`Failed to start process: ${err.toString()}`)
  }

  public mount(domElement: HTMLElement) {
    if (this.xterm) {
      this.xterm.dispose()
    }

    this.xterm = new Terminal({
      cursorBlink: true,
      cursorStyle: 'bar',
      convertEol: true,
    })

    this.fitAddon = new FitAddon()

    this.xterm.loadAddon(this.fitAddon)

    let command = ''

    this.xterm.onData(async (e) => {
      switch (e) {
        case '\u0003': // Ctrl+C
          // TODO send close signal
          break
        case '\u0004': // Ctrl+D
          await this.adapter?.write('^D\n')
          command = ''
          break
        case '\r': // Enter
          await this.adapter?.write(command + '\n')
          command = ''
          this.xterm?.write('\n')
          break
        case '\u007F': // Backspace (DEL)
          if (command.length > 0) {
            this.xterm?.write(ANSI_LEFT_CHAR_DELETE)
            command = command.substring(0, command.length - 1)
          }
          break
        default: // Print all other characters for demo
          if ((e >= String.fromCharCode(0x20) && e <= String.fromCharCode(0x7e)) || e >= '\u00a0') {
            command += e
            this.xterm?.write(e)
          }
      }
    })

    this.xterm.open(domElement)

    this.fitAddon.fit()
    this.xterm.focus()
  }

  public unmount() {
    this.xterm?.dispose()
  }
}
