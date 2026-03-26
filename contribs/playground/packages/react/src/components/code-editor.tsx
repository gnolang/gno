import React from 'react'

import { CMCodeEditor, type CMCodeEditorProps } from '@gnostudio/core'

export class CMCodeEditorReact extends React.Component<CMCodeEditorProps> {
  private readonly editor: CMCodeEditor
  private readonly containerRef = React.createRef<HTMLDivElement>()

  constructor(props: CMCodeEditorProps) {
    super(props)
    this.editor = new CMCodeEditor(props)
  }

  render() {
    return (
      <div
        ref={this.containerRef}
        className="code-editor"
        style={{ width: '100%', height: '100%', position: 'relative', overflow: 'hidden' }}
      />
    )
  }

  componentDidMount() {
    if (!this.containerRef.current) {
      return
    }

    this.editor.mount(this.containerRef.current)
  }

  componentWillUnmount() {
    this.editor.unmount()
  }

  componentDidUpdate() {
    this.editor.updateProps(this.props)
  }
}
