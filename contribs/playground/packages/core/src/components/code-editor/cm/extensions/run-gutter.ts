import { Facet, RangeSet, StateEffect, StateField, type EditorState, type Extension } from '@codemirror/state'
import { EditorView, gutter, GutterMarker, ViewPlugin } from '@codemirror/view'

import { derivePackageStructure, type FunctionSymbol } from '../../../../services'
import { debounce } from '../../../../utils/debounce'

export type onRunCallback = (payload: { func: FunctionSymbol; values: string[] }) => void

const runWidgetClassNames = {
  form: 'cm-run-widget__form',
  container: 'cm-run-widget',
  button: 'cm-run-widget__button',
  popover: {
    container: 'cm-run-widget__popover',
    content: 'cm-run-widget__popover-content',
  },
  input: 'cm-run-widget__input',
}

const getMarkerRanges = (state: EditorState) => {
  const config = state.facet(runMarkerConfig)
  const onRun = config[0]?.onRun

  let set: RangeSet<GutterMarker> = RangeSet.empty

  const doc = derivePackageStructure(state.doc.toString())
  doc.functions.forEach((func) => {
    const marker = new RunMarker(func, onRun)
    set = set.update({ add: [marker.range(func.pos.from, func.pos.to)] })
  })

  return set
}

class RunMarker extends GutterMarker {
  constructor(
    public options: FunctionSymbol,
    public onRun?: onRunCallback,
  ) {
    super()
    this.options = options
  }

  eq(other: RunMarker) {
    return this.options.pos.from === other.options.pos.from && this.options.pos.to === other.options.pos.to
  }

  toDOM() {
    const container = document.createElement('div')
    const button = document.createElement('div')
    const popover = document.createElement('div')
    const popoverContent = document.createElement('div')

    container.className = runWidgetClassNames.container
    button.className = runWidgetClassNames.button
    popover.className = runWidgetClassNames.popover.container
    popoverContent.className = runWidgetClassNames.popover.content

    button.role = 'button'
    button.title = 'Run function'
    popover.style.display = 'none'

    button.innerHTML = `<svg width="1em" height="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>`

    button.addEventListener('click', () => {
      // Show popover if function has parameters
      if (!this.options.params.length) {
        this.onRun?.({ func: this.options, values: [] })
        return
      }

      setTimeout(() => {
        popover.style.display = 'block'
        popover.querySelector('input')?.focus()
        document.body.addEventListener('click', closeOutsideListener)
        document.body.addEventListener('keydown', (e) => {
          if (e.key === 'Escape') {
            popover.style.display = 'none'
          }
        })
      }, 0)
    })

    // Close popover when clicking outside
    const closeOutsideListener = (e: MouseEvent) => {
      if ((e.target as HTMLElement).closest(runWidgetClassNames.popover.container)) return
      popover.style.display = 'none'
      document.body.removeEventListener('click', closeOutsideListener)
    }
    document.body.addEventListener('click', closeOutsideListener)

    const form = this.buildForm(popover)
    popoverContent.appendChild(form)

    popover.appendChild(popoverContent)
    container.appendChild(popover)
    container.appendChild(button)

    return container
  }

  private buildForm(popover: HTMLElement) {
    const form = document.createElement('form')
    form.className = runWidgetClassNames.form

    form.addEventListener('submit', (e) => {
      e.preventDefault()
      const formData = new FormData(form)
      const values = this.options.params.map((param) => {
        const value = formData.get(param.name ?? '')
        return value ? (value as string).toString() : ''
      })
      popover.style.display = 'none'
      this.onRun?.({ func: this.options, values })
    })

    for (const param of this.options.params) {
      const input = document.createElement('input')
      input.name = param.name ?? ''
      input.type = 'text'
      input.autocomplete = 'off'
      input.placeholder = `${param.name ?? 'value'} (${param.type ?? 'any'})`
      input.className = runWidgetClassNames.input
      form.appendChild(input)
    }

    const submit = document.createElement('input')
    submit.type = 'submit'
    submit.value = 'Run'

    form.appendChild(submit)

    return form
  }
}

interface RunMarkerConfig {
  onRun?: onRunCallback
}

const runMarkerConfig = Facet.define<RunMarkerConfig>()
const stateEffect = StateEffect.define<RangeSet<GutterMarker>>()
const stateField = StateField.define<RangeSet<GutterMarker>>({
  create(state) {
    return getMarkerRanges(state)
  },
  update(set, tr) {
    for (const effect of tr.effects) {
      if (effect.is(stateEffect)) {
        return effect.value
      }
    }
    return set
  },
})

const dispatchDecorations = debounce((view: EditorView) => {
  view.dispatch({
    effects: stateEffect.of(getMarkerRanges(view.state)),
  })
}, 1000)

/**
 * The ViewPlugin will watch for changes in the document
 * instead of the StateField, because we need to debounce the method
 * that retrieves the document structure to improve performance
 */
const viewPlugin = ViewPlugin.define(() => {
  return {
    update(viewUpdate) {
      if (viewUpdate.docChanged) {
        dispatchDecorations(viewUpdate.view)
      }
    },
  }
})

interface RunGutterOpts {
  onRun?: onRunCallback
}

export const runGutterExtension = (opts: RunGutterOpts): Extension => {
  return [
    runMarkerConfig.of({ onRun: opts.onRun }),
    viewPlugin,
    stateField,
    gutter({
      class: 'cm-run-gutter',
      markers: (view) => view.state.field(stateField),
    }),
    EditorView.baseTheme({
      '.cm-lineNumbers .cm-gutterElement': {
        'padding-left': '2px',
      },
      '.cm-run-gutter .cm-gutterElement': {
        padding: '0 0 0 4px',
      },
      '.cm-run-widget': {
        height: '100%',
        display: 'flex',
        'align-items': 'center',
      },
      '.cm-run-widget__button': {
        'font-size': '0.97rem',
        color: 'var(--colors-foreground-muted)',
        'user-select': 'none',
      },
      '.cm-run-widget__form input': {
        border: '1px solid var(--colors-border)',
        background: 'var(--colors-background)',
        color: 'var(--colors-foreground)',
      },
      '.cm-run-widget__form': {
        display: 'flex',
        'flex-direction': 'column',
        gap: '6px',
      },
      '.cm-run-widget__popover-content': {
        position: 'absolute',
        transform: 'translate(10px, 10px)',
        'background-color': 'var(--colors-background)',
        border: '1px solid var(--colors-border)',
        padding: '6px',
        borderRadius: '6px',
        'z-index': '1000',
      },
    }),
  ]
}
