import { LanguageServerClient } from '@gnostudio/codemirror-lsp'

import { EditorState, type Extension, type StateEffect } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { action, makeObservable, observable, reaction, toJS } from 'mobx'

import type { EditorEvent } from '../common/types'
import { CMEditorController } from './controller'
import { extendedSetup } from './extensions/extended-setup'
import { defaultPluginOptions, updateLspExtensionEffect } from './extensions/lsp'
import { readOnlyEffect } from './extensions/readonly'
import { EditorStateManager } from './manager'
import {
  checkStateDataChanges,
  getStateData,
  hasStateData,
  markDocumentDownloadedEffect,
  replaceStateFileName,
  setStateDataEffect,
  type StateData,
} from './state'
import type { EditorProps, File } from './types'
import { docFromString } from './utils/doc'
import {
  mapLspMessage,
  mapLspProgress,
  newFileLoadErrorEvent,
  newLspStartErrorMessage,
  newLspStartEvent,
  newTerminateEvent,
} from './utils/notification'

export type CMCodeEditorProps = EditorProps

/**
 * CMCodeEditor wraps and configures CodeMirror code editor.
 *
 * Editor state is done outside of React's state as we're managing editor instance directly and VDOM remains constant.
 */
export class CMCodeEditor {
  private readonly extensions: Extension[]
  private readonly stateMgr: EditorStateManager
  private remoteCtrl?: CMEditorController
  private lspClient?: LanguageServerClient
  private previousRaf?: number

  public view?: EditorView

  public props: EditorProps = {}

  constructor(props: EditorProps) {
    const { manager } = props

    this.props = props
    this.stateMgr = manager ?? new EditorStateManager()
    this.updateLspClientFromProps()

    this.extensions = extendedSetup({
      getState: () => this.props,
      onDocumentChange: (doc) => {
        const { fileName } = getStateData((this.view as EditorView).state)
        this.props.onChange?.(doc, fileName as string)
      },
      onGutterClick: (line) =>
        this.remoteCtrl?.onGutterClick?.({ startLineNumber: line, startColumn: 0, endLineNumber: line, endColumn: 0 }),
      onSave: (text) => this.props.onSave?.(text),
      onContextMenuAction: (e) => this.props.onContextMenuAction?.(e),
      onFormatDocument: () => this.remoteCtrl?.formatDocument(),
      onRunFunction: this.props.onRunFunction,
      onViewUpdate: (update) => this.remoteCtrl?.handleViewUpdate(update),
      showLineNumbers: props.showLineNumbers ?? true,
    })

    makeObservable(this, {
      props: observable,
      updateProps: action,
    })

    reaction(
      () => toJS(this.props),
      (_, prevProps) => this.updated(prevProps),
    )
  }

  updateProps(props: Partial<EditorProps>) {
    this.props = { ...this.props, ...props }
  }

  private getStateForFile(file?: File) {
    if (!file) {
      // Reset as no active document but still keep extensions state from previous state.
      const stateData = replaceStateFileName(this.view?.state, undefined)
      const state = this.getInitialEditorState()
      return { state, stateData, isCached: false, loadPromise: null }
    }

    const { path, content } = file
    const cachedState = this.stateMgr.getFileState(path)
    if (cachedState && hasStateData(cachedState)) {
      const stateData = getStateData(cachedState)
      return { state: cachedState, stateData, isCached: true, loadPromise: null }
    }

    // Lazy file content will be updated after loading is completed.
    const loadPromise: Promise<string> | null = typeof content === 'function' ? content(path) : null
    const doc = typeof content === 'string' ? docFromString(content) : ''

    const state = EditorState.create({
      doc,
      extensions: this.extensions,
    })

    // Extend document state data for a new file from a previous state.
    const stateData = replaceStateFileName(this.view?.state, path)
    stateData.isLoading = !!loadPromise
    return { state, stateData, loadPromise, isCached: false }
  }

  private getInitialEditorState() {
    return EditorState.create({
      extensions: this.extensions,
    })
  }

  mount(parent: HTMLElement) {
    const { value, lsp } = this.props

    let editorState: EditorState | undefined
    let isLazy = false
    if (value) {
      const { state, stateData, loadPromise } = this.getStateForFile(value)
      editorState = state

      if (loadPromise) {
        // initialize LSP and file contents later.
        isLazy = true
        this.waitForDocContents(stateData, loadPromise)
      }
    }

    this.view = new EditorView({
      parent,
      state: editorState,
    })

    const ctrl = new CMEditorController(this.view, this.props.formatter)
    this.props.onMount?.(ctrl)
    this.remoteCtrl = ctrl

    // If StateData field it not defined in EditorState, state will create
    // a new copy with current props every time it was queried instead of doing that just once per state.
    // This breaks change detection and to avoid that, we should explicitly set it every time.
    const effects: Array<StateEffect<any>> = [setStateDataEffect(this.props)]

    if (!isLazy) {
      // Initialize LSP client if possible.
      if (this.lspClient && lsp && value) {
        effects.push(
          updateLspExtensionEffect({
            // TODO: compute from state of props.
            ...defaultPluginOptions,
            languageId: 'gno',
            documentUri: `${lsp.rootUri}/${value.path}`,
            client: this.lspClient,
          }),
        )
      }
    } else {
      // Lock view until content is loaded.
      effects.push(readOnlyEffect(true))
    }

    this.view.dispatch({
      effects,
    })
  }

  private scheduleNotification(event: EditorEvent) {
    if (this.previousRaf) {
      cancelAnimationFrame(this.previousRaf)
    }

    this.previousRaf = requestAnimationFrame(() => {
      this.props?.onEvent?.(event)
    })
  }

  /**
   * Toggles LSP client on or off according to props changes.
   */
  private updateLspClientFromProps(prevProps?: Readonly<EditorProps>) {
    if (prevProps && prevProps.workspaceId !== this.props.workspaceId) {
      this.lspClient?.close()
      this.lspClient = undefined
      this.remoteCtrl?.setLspClient(null)
    }

    if (this.props.lsp && !this.lspClient) {
      // TODO: compare LSP config
      const { rootUri, transport, bootstrapHook, customNotifications } = this.props.lsp

      try {
        this.lspClient = new LanguageServerClient({
          rootUri,
          transport,
          bootstrapHook,
          customNotifications,
          windowNotificationHandler: {
            showMessage: (msg) => this.scheduleNotification(mapLspMessage(msg)),
            onProgress: (msg) => this.scheduleNotification(mapLspProgress(msg)),
            onInitBegin: () => {
              this.scheduleNotification(newLspStartEvent())
            },
            onInitError: (err) => {
              console.error('Failed to start lsp server:', err)
              this.scheduleNotification(newLspStartErrorMessage())
            },
          },
        })

        this.remoteCtrl?.setLspClient(this.lspClient)
      } catch (err) {
        console.error(err)
        this.scheduleNotification(newLspStartErrorMessage())
      }
      return
    }

    if (!this.props.lsp) {
      this.lspClient?.close()
      this.lspClient = undefined

      // Clean any pending statues
      this.props?.onEvent?.(newTerminateEvent())
    }
  }

  private waitForDocContents(stateData: StateData, promise: Promise<string>) {
    const { workspaceId } = this.props
    const { fileName } = stateData
    if (!fileName) {
      // Catch logical bugs.
      throw new Error('cannot lazy load contents with empty file name in state data.')
    }

    this.view?.dispatch({
      effects: [readOnlyEffect(true)],
    })

    type CallbackArgs = { content: null; err: Error } | { content: string; err?: undefined }
    const handler = ({ content, err }: CallbackArgs) => {
      if (!this.view) {
        return
      }

      // Skip update if file changed before previous content was fetched.
      // Covers error events as well to avoid showing outdated errors.
      const currentWorkspaceId = this.props.workspaceId
      if (currentWorkspaceId !== workspaceId) {
        return
      }

      const { fileName: currentFileName, isInitialised } = getStateData(this.view.state)

      if (!isInitialised || currentFileName !== fileName) {
        return
      }

      if (err) {
        this.props?.onEvent?.(newFileLoadErrorEvent(workspaceId ?? '', fileName, err))
        return
      }

      const effects = markDocumentDownloadedEffect(
        {
          lspClient: this.lspClient,
        },
        this.props,
      )

      // First, update contents and then start plugins.
      this.view.dispatch(
        {
          changes: {
            from: 0,
            to: this.view.state.doc.length,
            insert: content,
          },
        },
        {
          effects,
        },
      )
    }

    promise.then((content) => handler({ content })).catch((err) => handler({ content: null, err }))
  }

  /**
   * Applies updates on EditorState based on prop changes.
   */
  updated(prevProps: Readonly<EditorProps>) {
    if (!this.view) {
      // We can't do much without editor instance or when editor is loading required resources.
      return
    }
    let { value: prevFile } = prevProps
    const { value: newFile, formatter } = this.props

    this.updateLspClientFromProps(prevProps)
    this.remoteCtrl?.setFormatter(formatter)
    const workspaceChanged = prevProps.workspaceId !== this.props.workspaceId
    if (workspaceChanged) {
      this.stateMgr.clear()

      // Remove reference to avoid referencing detached MobX tree leaf.
      prevFile = undefined
    }

    const fileChanged = prevFile?.path !== newFile?.path

    if (!workspaceChanged && fileChanged && prevFile) {
      const { state } = this.view
      // Ignore persist of previous state, if it's a default editor state (not from a file) or not file loaded yet.
      // E.g. when made using getInitialEditorState().
      const { isInitialised, isLoading } = getStateData(this.view.state)
      if (isInitialised && !isLoading) {
        this.stateMgr.replaceFileState(prevFile.path, state)
      }
    }

    let currentStateData: StateData
    if (fileChanged) {
      // getStateForFile creates a new, blank state for documents that weren't previously cached.
      // That means it should be called only during document switch.
      // Otherwise - change detector will always compare against a default state, triggering constant updates.
      const { state, stateData, loadPromise } = this.getStateForFile(newFile)
      currentStateData = stateData

      // Effects can be dispatched only on mounted
      this.view.setState(state)

      if (loadPromise) {
        // New file with lazy-fetch content
        this.waitForDocContents(stateData, loadPromise)
      }
    } else {
      // Compare current state if document is not changed.
      currentStateData = getStateData(this.view.state)
    }

    const ctx = {
      lspClient: this.lspClient,
    }
    const { effects, changes, isChanged } = checkStateDataChanges(ctx, this.props, currentStateData)
    if (!isChanged) {
      return
    }

    this.view.dispatch({ effects })
    this.remoteCtrl?.checkInputModeChanges(changes)

    if (fileChanged) {
      // Sync state cursor to status bar.
      this.remoteCtrl?.broadcastCursorState?.(this.view?.state)
      this.view?.focus()
    }
  }

  unmount() {
    this.remoteCtrl?.setLspClient(null)
    this.props?.onEvent?.(newTerminateEvent())
    this.stateMgr.clear()
    this.remoteCtrl?.dispose()
    this.view?.destroy()
  }
}
